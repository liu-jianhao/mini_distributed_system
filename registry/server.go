package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":8020"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []*Registration
	mutex         *sync.RWMutex
}

func (r *registry) add(reg *Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()

	return nil
}

var reg = registry{
	registrations: make([]*Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{}

func (rs *RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var registration Registration
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&registration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Adding Service name=%v, url=%s", registration.ServiceName, registration.ServiceURL)

		err = reg.add(&registration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
