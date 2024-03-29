package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"

	"github.com/google/uuid"

	"github.com/iharsuvorau/garlic/instruction"
)

type Images struct {
	Images []*instruction.ShowImage

	filepath string
	mu       sync.RWMutex
}

func NewImageStore(fpath string) (*Images, error) {
	var file *os.File
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		file, err = os.Create(fpath)
		if err != nil {
			return nil, fmt.Errorf("can't create an image store at %s: %v", fpath, err)
		}
	} else {
		file, err = os.Open(fpath)
	}
	defer file.Close()

	store := &Images{filepath: fpath}
	if err = json.NewDecoder(file).Decode(&store.Images); err != nil && err != io.EOF {
		return nil, fmt.Errorf("can't decode images from %s: %v", fpath, err)
	}

	return store, store.dump()
}

func (s *Images) GetByUUID(id uuid.UUID) (*instruction.ShowImage, error) {
	for _, v := range s.Images {
		if v.ID == id {
			return v, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

//func (s *Images) GetByName(name string) (*instruction.ShowImage, error) {
//	for _, v := range s.Images {
//		if v.Name == name {
//			return v, nil
//		}
//	}
//
//	return nil, fmt.Errorf("not found")
//}

func (s *Images) Get(id string) (*instruction.ShowImage, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	for _, v := range s.Images {
		if v.ID == uid {
			return v, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (s *Images) Create(v *instruction.ShowImage) error {
	if (v.ID == uuid.UUID{}) {
		return fmt.Errorf("failed to create an image: ID must be provided")
	}

	//if _, err := s.GetByName(v.Name); err == nil {
	//	return fmt.Errorf("the image with such name already exists: %v", v.Name)
	//}

	s.mu.Lock()
	s.Images = append(s.Images, v)
	s.mu.Unlock()
	return s.dump()
}

func (s *Images) Update(updatedImage *instruction.ShowImage) error {
	s.mu.Lock()
	for _, s := range s.Images {
		if s.ID == updatedImage.ID {
			*s = *updatedImage
		}
	}
	s.mu.Unlock()

	return s.dump()
}

func (s *Images) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	item, err := s.Get(id)
	if err != nil {
		return err
	}

	newItems := []*instruction.ShowImage{}

	for _, s := range s.Images {
		if s.ID == uid {
			continue
		}
		newItems = append(newItems, s)
	}

	s.mu.Lock()
	if err = os.Remove(item.FilePath); err != nil {
		return fmt.Errorf("failed to remove a file: %v", err)
	}
	s.Images = newItems
	s.mu.Unlock()

	return s.dump()
}

func (s *Images) GetGroups() []string {
	var groupsMap = map[string]interface{}{}

	for _, v := range s.Images {
		if v == nil {
			continue
		}

		groupsMap[v.Group] = nil
	}

	var groups = make([]string, len(groupsMap))
	var i int64 = 0
	for k := range groupsMap {
		groups[i] = k
		i++
	}
	sort.Strings(groups)

	return groups
}

func (s *Images) dump() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Create(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to open a file: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(s.Images)
}
