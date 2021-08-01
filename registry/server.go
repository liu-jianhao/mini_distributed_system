package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	err := r.sendRequiredServices(reg)
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) remove(url string) error {
	for i, registration := range reg.registrations {
		if registration.ServiceURL == url {
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

func (r *registry) sendRequiredServices(reg *Registration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	p := patch{}
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName != reqService {
				continue
			}
			p.Added = append(p.Added, &patchEntry{
				Name: serviceReg.ServiceName,
				URL:  serviceReg.ServiceURL,
			})
		}
	}

	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}

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
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
