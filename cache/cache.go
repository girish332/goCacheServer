package cache

import (
	"cacheServer/appcontext"
	"cacheServer/apperror"
	"github.com/lib/pq"
	"log"
	"os"
	"runtime"
	"sort"
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
	GetCategoryIndicesCache() ([]int, error)
	GetMaximumIndexCategory() (int, error)
	GetMaximumIndexSubcategory(categoryID string) (int, error)
	UpdateCategoryIndexCache(index int) error
	UpdateSubcategoryIndexCache(index int, categoryID string) error
	DeleteCategoryIndexCache(key int) error
	DeleteSubcategoryIndexCache(categoryID string, index int) error
	GetSubcategoryIndicesCache(categoryID string) ([]int, error)
	CreateSubcategoryCache(categoryID string) error
	DeleteProductCacheIndex(subcategoryID string, key int) error
	GetProductIndicesCache(subcategoryID string) ([]int, error)
	CreateProductCache(subcategoryID string) error
	GetMaximumIndexProduct(subcategoryID string) (int, error)
	UpdateProductCacheIndex(index int, subcategoryID string) error
}

// Request ...
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
	go s.initializeCategoryCache()
	go s.initializeSubcategoryCache()
	go s.initializeProductCache()
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

// GetCategoryIndicesCache ...
func (s *Server) GetCategoryIndicesCache() ([]int, error) {
	var result []int
	for k, v := range s.store.categoryIndices {
		if v == true {
			result = append(result, k)
		}
	}
	if len(result) == 0 {
		err := s.initializeCategoryCache()
		if err != nil {
			return nil, apperror.ErrCacheNotInitialized
		}
	}

	return result, nil
}

// GetSubcategoryIndicesCache ...
func (s *Server) GetSubcategoryIndicesCache(categoryID string) ([]int, error) {
	var result []int
	for k, v := range s.store.subcategoryIndices[categoryID] {
		if v == true {
			result = append(result, k)
		}
	}
	if len(result) == 0 {
		err := s.initializeSubcategoryCache()
		if err != nil {
			return nil, apperror.ErrCacheNotInitialized
		}
	}
	return result, nil
}

// CreateSubcategoryCache ...
func (s *Server) CreateSubcategoryCache(categoryID string) error {
	s.store.Lock()
	var result [255]bool
	result[1] = true
	s.store.subcategoryIndices[categoryID] = result
	s.store.Unlock()
	return nil
}

// GetMaximumIndexCategory : returns the maximum value present in availableIndices
func (s *Server) GetMaximumIndexCategory() (int, error) {
	arr, err := s.GetCategoryIndicesCache()
	if err != nil {
		return 0, err
	}
	return arr[len(arr)-1], nil
}

// GetMaximumIndexSubcategory : returns max subcategory index
func (s *Server) GetMaximumIndexSubcategory(categoryID string) (int, error) {
	arr, err := s.GetSubcategoryIndicesCache(categoryID)
	if err != nil {
		return 0, err
	}
	return arr[len(arr)-1], nil
}

// UpdateCategoryIndexCache ...
func (s *Server) UpdateCategoryIndexCache(index int) error {
	_, err := s.GetCategoryIndicesCache()
	if err != nil {
		return err
	}
	s.store.Lock()
	s.store.categoryIndices[index] = true
	s.store.Unlock()
	return nil
}

// DeleteCategoryIndexCache ...
func (s *Server) DeleteCategoryIndexCache(key int) error {
	_, err := s.GetCategoryIndicesCache()
	if err != nil {
		return err
	}
	s.store.Lock()
	s.store.categoryIndices[key] = false
	s.store.Unlock()
	return nil
}

// UpdateSubcategoryIndexCache ...
func (s *Server) UpdateSubcategoryIndexCache(index int, categoryID string) error {
	_, err := s.GetSubcategoryIndicesCache(categoryID)
	if err != nil {
		return err
	}
	s.store.Lock()
	res := s.store.subcategoryIndices[categoryID]
	res[index] = true
	s.store.subcategoryIndices[categoryID] = res
	s.store.Unlock()
	return nil
}

// DeleteSubcategoryIndexCache ...
func (s *Server) DeleteSubcategoryIndexCache(categoryID string, index int) error {
	_, err := s.GetSubcategoryIndicesCache(categoryID)
	if err != nil {
		return err
	}
	s.store.Lock()
	res := s.store.subcategoryIndices[categoryID]
	res[index] = false
	s.store.subcategoryIndices[categoryID] = res
	s.store.Unlock()
	return nil
}

type occupiedSubcategoryIndices struct {
	categoryID string
	indices    []int32
}

func (s *Server) initializeSubcategoryCache() error {
	query := `SELECT "categoryID",ARRAY_AGG("index") FROM (
              SELECT "categoryID","index" FROM "productSubCategory" GROUP BY 1,2 ORDER BY 2 ASC) t1 
              GROUP BY 1;`
	result, err := s.appCtx.DatabaseClient.Query(query)
	if err != nil {
		return err
	}

	for result.Next() {
		var subcategoryIndex occupiedSubcategoryIndices
		err := result.Scan(&subcategoryIndex.categoryID, (*pq.Int32Array)(&subcategoryIndex.indices))
		if err != nil {
			log.Println("failed to initialize subcategory cache", err)
			return err
		}
		s.store.subcategoryIndices[subcategoryIndex.categoryID] = s.getMissingIndices(subcategoryIndex.indices)
	}
	return nil
}

func (s *Server) getMissingIndices(occupiedIndices []int32) [255]bool {
	var result [255]bool
	var count int32
	count = 1
	var maxValue int32
	maxValue = 0
	for _, v := range occupiedIndices {
		if count != v {
			for count < v {
				result[int(count)] = true
				count++
			}
		}
		count++
		maxValue = v
	}
	result[maxValue+1] = true

	return result
}

// initializeCategoryCache ...
func (s *Server) initializeCategoryCache() error {
	query := `SELECT index from "productCategory" ORDER BY index ASC;`
	result, err := s.appCtx.DatabaseClient.Query(query)
	if err != nil {
		return err
	}

	count := 1
	var index int
	for result.Next() {
		if err = result.Scan(&index); err != nil {
			log.Println("failed to initialize category cache", err)
			return err
		}
		if count != index {
			for count < index {
				s.store.categoryIndices[count] = true
				count++
			}
		}
		count++
	}
	s.store.categoryIndices[index+1] = true
	return nil
}

// GetProductIndicesCache ...
func (s *Server) GetProductIndicesCache(subcategoryID string) ([]int, error) {
	res := s.store.productIndices[subcategoryID].orderedIndices
	if len(res) == 0 {
		err := s.initializeProductCache()
		if err != nil {
			return nil, apperror.ErrCacheNotInitialized
		}
	}

	return res, nil
}

// CreateProductCache ...
func (s *Server) CreateProductCache(subcategoryID string) error {
	s.store.Lock()
	//orderedIndices.store.ProductIndex[subcategoryID] = []int64{1}
	p := NewSortedIndices(make([]int, 0))
	p.orderedIndices = append(p.orderedIndices, 1)
	p.availableIndices[1] = struct{}{}
	s.store.productIndices[subcategoryID] = p
	s.store.Unlock()
	return nil
}

// UpdateProductCacheIndex ...
func (s *Server) UpdateProductCacheIndex(index int, subcategoryID string) error {
	_, err := s.GetProductIndicesCache(subcategoryID)
	if err != nil {
		return err
	}
	s.store.Lock()
	s.store.productIndices[subcategoryID].availableIndices[index] = struct{}{}
	go func() {
		// sorting the indices in a goroutine to not affect performance of end user
		var result []int
		for k := range s.store.productIndices[subcategoryID].availableIndices {
			result = append(result, k)
		}
		sort.Ints(result)
		s.store.productIndices[subcategoryID].orderedIndices = result
	}()
	s.store.Unlock()
	return nil
}

// DeleteProductCacheIndex ...
func (s *Server) DeleteProductCacheIndex(subcategoryID string, index int) error {
	_, err := s.GetProductIndicesCache(subcategoryID)
	if err != nil {
		return err
	}
	s.store.Lock()
	delete(s.store.productIndices[subcategoryID].availableIndices, index)
	go func() {
		// sorting the indices in a goroutine to not affect performance of end user
		var result []int
		for k := range s.store.productIndices[subcategoryID].availableIndices {
			result = append(result, k)
		}
		sort.Ints(result)
		s.store.productIndices[subcategoryID].orderedIndices = result
	}()
	s.store.Unlock()
	return nil
}

// GetMaximumIndexProduct ...
func (s *Server) GetMaximumIndexProduct(subcategoryID string) (int, error) {
	result, err := s.GetProductIndicesCache(subcategoryID)
	if err != nil {
		return 0, err
	}
	return result[len(result)-1], nil
}

// fillAvailableIndices uses the indices which are occupied to get the indices which are available
// and stores them in cache
func (s *Server) fillAvailableIndices(subcategoryID string, occupiedIndices []int32) {
	var count int32
	count = 1
	var maxValue int32
	maxValue = 0
	p := NewSortedIndices(make([]int, 0))
	s.store.productIndices[subcategoryID] = p
	for _, v := range occupiedIndices {
		if count != v {
			for count < v {
				s.store.productIndices[subcategoryID].availableIndices[int(count)] = struct{}{}
				s.store.productIndices[subcategoryID].orderedIndices = append(s.store.productIndices[subcategoryID].orderedIndices, int(count))
				count++
			}
		}
		count++
		maxValue = v
	}

	s.store.productIndices[subcategoryID].availableIndices[int(maxValue)+1] = struct{}{}
	s.store.productIndices[subcategoryID].orderedIndices = append(s.store.productIndices[subcategoryID].orderedIndices, int(maxValue)+1)
	return
}

type occupiedIndices struct {
	subcategoryID string
	indices       []int32
}

// initializeProductCache ...
func (s *Server) initializeProductCache() error {

	query2 := `SELECT id FROM "productSubCategory";`
	result, err := s.appCtx.DatabaseClient.Query(query2)
	if err != nil {
		log.Println("Error getting subcategoryID :", err)
		return err
	}
	for result.Next() {
		var subcategoryID string
		err := result.Scan(&subcategoryID)
		if err != nil {
			log.Println("Error scanning subcategoryID:", err)
			return err
		}
		p := NewSortedIndices(make([]int, 0))
		s.store.productIndices[subcategoryID] = p
		s.store.productIndices[subcategoryID].availableIndices[1] = struct{}{}
		s.store.productIndices[subcategoryID].orderedIndices = append(s.store.productIndices[subcategoryID].orderedIndices, 1)
	}

	query := `SELECT "subCategoryID",ARRAY_AGG("index") FROM (
              SELECT "subCategoryID","index" FROM "products" GROUP BY 1,2 ORDER BY 2 ASC) t1 
              GROUP BY 1;`
	result, err = s.appCtx.DatabaseClient.Query(query)
	if err != nil {
		log.Println(err)
		return err
	}

	var count int
	for result.Next() {
		var productIndex occupiedIndices
		err := result.Scan(&productIndex.subcategoryID, (*pq.Int32Array)(&productIndex.indices))
		if err != nil {
			return err
		}
		count++
		s.fillAvailableIndices(productIndex.subcategoryID, productIndex.indices)
	}
	return nil
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
