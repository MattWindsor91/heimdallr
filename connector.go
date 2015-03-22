package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/UniversityRadioYork/baps3-go"
)

// GetResponse is the outer structure of all GET responses.
type GetResponse struct {
	Status string
	Value  interface{}
}

// GetOk creates a GetResponse wrapping a successful GET result.
func GetOk(value interface{}) *GetResponse {
	r := new(GetResponse)
	r.Status = "ok"
	r.Value = value
	return r
}

type bfConnector struct {
	conn   *baps3.Connector
	name   string
	wg     *sync.WaitGroup
	logger *log.Logger
	state  *serviceState

	reqCh chan httpRequest
	resCh <-chan baps3.Message

	// TODO(CaptainHayashi): move this away from baps3.Message to
	// something generic.
	updateCh chan<- baps3.Message
}

func initBfConnector(name string, updateCh chan baps3.Message, waitGroup *sync.WaitGroup, logger *log.Logger) (c *bfConnector) {
	resCh := make(chan baps3.Message)

	c = new(bfConnector)
	c.resCh = resCh
	c.conn = baps3.InitConnector(name, resCh, waitGroup, logger)
	c.name = name
	c.wg = waitGroup
	c.logger = logger
	c.reqCh = make(chan httpRequest)
	c.updateCh = updateCh
	c.state = initServiceState()
	return
}

func (c *bfConnector) Run() {
	defer c.wg.Done()
	defer close(c.conn.ReqCh)

	go c.conn.Run()

	fmt.Printf("connector %s now listening for requests\n", c.name)

	for {
		select {
		case rq, ok := <-c.reqCh:
			if !ok {
				return
			}

			// TODO(CaptainHayashi): probably make this more robust
			resource := strings.Replace(rq.resource, "/"+c.name, "", 1)
			fmt.Printf("connector %s response %s\n", c.name, resource)

			// TODO(CaptainHayashi): other methods
			rq.resCh <- c.get(resource)
		case res := <-c.resCh:
			if err := c.state.update(res); err != nil {
				fmt.Println(err)
			}
			c.updateCh <- res
		}
	}

	return
}

func splitResource(resource string) []string {
	res := strings.Split(strings.Trim(resource, "/"), "/")

	// The empty resource is returned as {""}: let's fix that
	if len(res) == 1 && res[0] == "" {
		res = []string{}
	}

	return res
}

func (c *bfConnector) get(resource string) interface{} {
	// TODO(CaptainHayashi): HTTP status codes

	resourcePath := splitResource(resource)

	r := c.rootGet(resourcePath)

	if r == nil {
		// TODO(CaptainHayashi): more errors
		return GetResponse{
			Status: "what",
			Value:  "resource not found: " + resource,
		}
	}

	return GetOk(r)
}

/* These resMaps describe simple composite resources, mapping each child
 * resource to the functions handling them.
 *
 * TODO(CaptainHayashi): add support for things that aren't GET
 *   (have each resource be a jump table of possible methods, or send the method
 *    to the resMap func?)
 * TODO(CaptainHayashi): decouple traversal from GET
 * TODO(CaptainHayashi): maybe make traversal iterative instead of recursive?
 */

type resMap map[string]func(*bfConnector, []string) interface{}

var (
	rootRes = resMap{
		"control": (*bfConnector).controlGet,
		"player":  (*bfConnector).playerGet,
		// "playlist": (*bfConnector).playlistGet
	}
	controlRes = resMap{
		"features": (*bfConnector).featuresGet,
		"state":    (*bfConnector).stateGet,
	}
	playerRes = resMap{
		"time": (*bfConnector).timeGet,
		"file": (*bfConnector).fileGet,
	}
)

func (c *bfConnector) getResource(rm resMap, resourcePath []string) interface{} {
	if len(resourcePath) == 0 {
		// Pull down all of the available child resources in this
		// resource.
		object := make(map[string]interface{})

		for k := range rm {
			child := rm[k](c, []string{})

			// Only add a key if the child definitely exists.
			if child != nil {
				object[k] = child
			}
		}

		return object
	}

	// Does the next step on the resource path exist?
	rfunc, ok := rm[resourcePath[0]]
	if ok {
		// Make it that resource's responsibility to
		// find the resource, then.
		return rfunc(c, resourcePath[1:])
	}
	return nil
}

// controlGet is the GET handler for the /control resource.
func (c *bfConnector) controlGet(resourcePath []string) interface{} {
	return c.getResource(controlRes, resourcePath)
}

// playerGet is the GET handler for the /player resource.
func (c *bfConnector) playerGet(resourcePath []string) interface{} {
	// TODO(CaptainHayashi): Probably a spec change, but the fact that this
	// resource is guarded by more than one feature is iffy.  Do we need a
	// Player feature?
	if !(c.state.hasFeature(FtFileLoad) || c.state.hasFeature(FtTimeReport)) {
		return nil
	}

	return c.getResource(playerRes, resourcePath)
}

// rootGet is the GET handler for the / resource.
func (c *bfConnector) rootGet(resourcePath []string) interface{} {
	return c.getResource(rootRes, resourcePath)
}

// GET value for /control/features
func (c *bfConnector) featuresGet(resourcePath []string) interface{} {
	// We only want a resource length of 0 (all features), or 1
	// (some index into the list of resources).
	if 1 < len(resourcePath) {
		return nil
	}

	fstrings := []string{}

	for k := range c.state.features {
		fstrings = append(fstrings, k.String())
	}

	// There's no need to sort these, but it doesn't hurt to make the list a
	// bit easier for humans to eyeball.
	sort.Strings(fstrings)

	// Did we want the whole resource (the features list)?
	if len(resourcePath) == 0 {
		return fstrings
	}

	// If not, we assume we were indexing into the features list.
	// TODO(CaptainHayashi): Factor this list-resource pattern out?
	//   It might be useful for playlists later.
	i, err := strconv.Atoi(resourcePath[0])
	// TODO(CaptainHayashi): handle err properly
	if err == nil && 0 <= i && i <= len(fstrings) {
		return fstrings[i]
	}

	return nil
}

// GET value for /control/state
func (c *bfConnector) stateGet(resourcePath []string) interface{} {
	if 0 < len(resourcePath) {
		return nil
	}

	return c.state
}

// GET value for /player/time
func (c *bfConnector) timeGet(resourcePath []string) interface{} {
	if 0 < len(resourcePath) {
		return nil
	}
	if !c.state.hasFeature(FtTimeReport) {
		return nil
	}

	// Time is reported in _micro_seconds
	return c.state.time.Nanoseconds() / 1000
}

// GET value for /player/file
func (c *bfConnector) fileGet(resourcePath []string) interface{} {
	if 0 < len(resourcePath) {
		return nil
	}
	if !c.state.hasFeature(FtFileLoad) {
		return nil
	}

	return c.state.file
}
