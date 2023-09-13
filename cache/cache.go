package cache

import (
	"cacheServer/appcontext"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// Type ...
type Type uint32

const (
	// Role ...
	Role Type = iota
	// Product ...
	Product
	// Category ...
	Category
	// Subcategory ...
	Subcategory
	// Quit ...
	Quit
)

func (t Type) String() string {
	return [...]string{"Role", "Product", "Category", "SubCategory"}[t]
}

// AppCache ...
type AppCache interface {
	MakeRequest(request *Request)
	DeleteCache(id string, t Type)
}

// Request struct
type Request struct {
	id      string
	reqType Type
	Out     chan bool   // Channel used to receive data from cache
	opt     interface{} // optional parameter
}

// NewRequest ...
func NewRequest(id string, Type Type, opt interface{}) *Request {
	return &Request{
		id:      id,
		reqType: Type,
		Out:     make(chan bool),
		opt:     opt,
	}
}

// Server ...
type Server struct {
	request chan Request
	store   Store
	appCtx  *appcontext.Context
}

// Store : map of maps (category,subcategory,product,role)
type Store struct {
	data               map[Type]map[string]string
	categoryIndices    [255]bool
	subcategoryIndices map[string][255]bool
	productIndices     map[string]*SortedIndices // subcategoryID vs struct
	sync.Mutex
}

func newStore() Store {
	return Store{
		data: map[Type]map[string]string{
			Role:        make(map[string]string), // Map of role id vs role
			Category:    make(map[string]string), // Map categoryID vs active/passive
			Subcategory: make(map[string]string), // Map subcategoryID vs active/passive
			Product:     make(map[string]string), // Map productID vs  active/passive
		},
		categoryIndices:    [255]bool{},
		subcategoryIndices: make(map[string][255]bool), // Map of categoryID vs availableIndices
		productIndices:     make(map[string]*SortedIndices),
	}
}

func newServer(appCtx *appcontext.Context) *Server {
	s := &Server{
		request: make(chan Request),
		store:   newStore(),
	}
	s.setAppCtx(appCtx)
	return s
}

// SortedIndices : struct used to store available indices and maintain the order
type SortedIndices struct {
	availableIndices map[int]struct{}
	orderedIndices   []int
}

// NewSortedIndices ...
func NewSortedIndices(s []int) *SortedIndices {
	return &SortedIndices{availableIndices: make(map[int]struct{}), orderedIndices: s}
}

var once sync.Once
var instance *Server

// GetCacheInstance : initializes singleton object of server
func GetCacheInstance(appCtx *appcontext.Context) *Server {
	once.Do(func() {
		log.Println("server instance initialized")
		instance = newServer(appCtx)
	})
	return instance
}

func (s *Server) setAppCtx(appCtx *appcontext.Context) {
	s.appCtx = appCtx
}

// Run ....
func (s *Server) Run() {
	maxProc, _ := strconv.Atoi(os.Getenv("GO_MAX_PROC"))
	runtime.GOMAXPROCS(maxProc)
	for {
		req := <-s.request
		switch req.reqType {
		case Role:
			log.Println("Request received for role verification")
			go s.verifyRequest(req, Role, true, "users")
		case Category:
			log.Println("Request received for categoryID verification")
			go s.verifyRequest(req, Category, false, "productCategory")
		case Subcategory:
			log.Println("Request received for subcategoryID verification")
			go s.verifyRequest(req, Subcategory, false, "productSubCategory")
		case Product:
			log.Println("Request received for productID verification")
			go s.verifyRequest(req, Product, false, "products")
		case Quit:
			return
		default:
			log.Println("Not supported")
		}
	}
}

// MakeRequest ....
func (s *Server) MakeRequest(request *Request) {
	s.request <- *request
}

// Close ...
func (s *Server) Close() {
	s.request <- *NewRequest("Quit", Quit, nil)
}

// verifyRequest : verifies the request from cache
func (s *Server) verifyRequest(req Request, reqType Type, isOpt bool, tableName string) {
	s.store.Lock()
	if cachedValue, ok := s.store.data[reqType][req.id]; ok {
		s.store.Unlock()
		if !isOpt { // isOpt is false for category,subcategory,product
			if req.opt == nil {
				req.Out <- "active" == cachedValue
				log.Println(reqType, "id fetched from cache")
			}
		} else {
			if req.opt != nil {
				// opt.(string) contains claimedRole from claims
				req.Out <- req.opt.(string) == cachedValue
				log.Println(reqType, "id fetched from cache")
			} else {
				req.Out <- false
				log.Println("isOpt not passed when required")
			}
		}
	} else {
		log.Println(reqType, " not present in cache")
		s.store.Unlock()
		// if not present in cache, fetch from db and update cache
		dbVal, err := s.fetchQuery(req.id, tableName, reqType)
		if err != nil {
			req.Out <- false
			return
		}

		go s.updateCache(dbVal, req.id, reqType)
		if isOpt {
			req.Out <- req.opt.(string) == dbVal
		} else {
			req.Out <- "active" == dbVal
		}
	}
}

func (s *Server) updateCache(dbVal string, id string, t Type) {
	s.store.Lock()
	s.store.data[t][id] = dbVal
	s.store.Unlock()
	log.Println("cache is updated")
}

// DeleteCache : pass in the id and the type to delete value in cache
func (s *Server) DeleteCache(id string, t Type) {
	s.store.Lock()
	delete(s.store.data[t], id)
	s.store.Unlock()
}

func (s *Server) fetchQuery(ID string, tableName string, t Type) (string, error) {
	if t != Role {
		query := `SELECT id FROM ` + `"` + tableName + `"` + ` WHERE id=$1;`
		result := s.appCtx.DatabaseClient.QueryRow(query, ID)
		var categoryID string
		err := result.Scan(&categoryID)
		if err != nil {
			log.Println("error while scanning db result ", err)
			return "passive", err
		}
		return "active", nil
	}
	query := `SELECT "role" FROM "users" WHERE "emailId" = $1`
	result := s.appCtx.DatabaseClient.QueryRow(query, ID)
	var dbRole string
	err := result.Scan(&dbRole)
	if err != nil {
		log.Println("error while scanning db result ", err)
		return "", err
	}
	return dbRole, nil
}
