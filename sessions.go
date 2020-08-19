package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// sessions represent all sessions available and is the main type for a user which simplifies
// the control of a Pepper robot.
var sessions []*Session

// moves represent all ready-made moves located somewhere on the disk. This presented in the
// web UI as a library of moves which can be called any time by a user.
var moves Moves

// moveGroups is a helper variable for the "/sessions/" route to list moves by a group.
var moveGroups []string

func collectSessions(sayDir string, moves *Moves) ([]*Session, error) {
	var sessions = []*Session{
		{
			Name: "Session 1",
			Items: []*SessionItem{
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Tere, mina olen robot Pepper. Mina olen 6-aastane ja tahan sinuga tuttavaks saada. Mis on sinu nimi?",
								FilePath: "1out_tutvustus.wav",
							},
							MoveItem: &MoveAction{
								Name:  "hello_a010",
								Delay: 0,
							},
						},
						{
							SayItem: &SayAction{
								Phrase: "Very nice",
							},
							MoveItem: &MoveAction{
								Name: "NiceReaction_01",
							},
						},
						{
							SayItem: &SayAction{
								Phrase: "That is sad",
							},
							MoveItem: &MoveAction{
								Name: "SadReaction_01",
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Kui vana sa oled?",
								FilePath: "2out_vanus.wav",
							},
							MoveItem: &MoveAction{
								Name:  "question_right_hand_a001",
								Delay: 0,
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Kas Sul on vendi või õdesid?",
								FilePath: "3out_vennad.wav",
							},
							MoveItem: &MoveAction{
								Name:  "question_both_hands_a007",
								Delay: 0,
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Ma tulin siia üksi, kuid mu pere on suur ja mööda maailma laiali.",
								FilePath: "3out_vennadVV.wav",
							},
							MoveItem: &MoveAction{
								Name:  "both_hands_high_b001",
								Delay: 0,
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Mina olen pärit Pariisist ja nüüd meeldib mulle väga Eestis elada. Mis sulle Sinu Eestimaa juures meeldib?",
								FilePath: "4out_päritolu.wav",
							},
							MoveItem: &MoveAction{
								Name:  "exclamation_both_hands_a001",
								Delay: time.Second * 5,
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase:   "Jaa, see on väike ja sõbralik maa ja teil on 4 aastaaega",
								FilePath: "5out_eestimaavastus.wav",
							},
							MoveItem: &MoveAction{
								Name:  "affirmation_a009",
								Delay: 0,
							},
						},
					},
				},
			},
		},
		{
			ID:   uuid.Must(uuid.NewRandom()),
			Name: "Session 2",
			Items: []*SessionItem{
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase: "Q1",
							},
						},
					},
				},
				{
					Actions: []*SayAndMoveAction{
						{
							SayItem: &SayAction{
								Phrase: "Q2",
							},
						},
					},
				},
			},
		},
	}

	for _, s := range sessions {
		s.ID = uuid.Must(uuid.NewRandom())

		for _, item := range s.Items {
			item.ID = uuid.Must(uuid.NewRandom()) // TODO: do we need ID for an item

			for _, action := range item.Actions {
				if action == nil {
					continue
				}

				// setting IDs
				action.SetID(uuid.Must(uuid.NewRandom()))
				if !action.IsNil() {
					action.SayItem.SetID(uuid.Must(uuid.NewRandom()))
					action.MoveItem.SetID(uuid.Must(uuid.NewRandom()))
				}

				// looking for audio files
				if action.SayItem.FilePath != "" {
					action.SayItem.FilePath = path.Join(sayDir, s.Name, action.SayItem.FilePath)
					if _, err := os.Stat(action.SayItem.FilePath); os.IsNotExist(err) {
						return nil, err
					}
				}

				// looking for moves
				if action.MoveItem != nil {
					// the move is found in the server's library, otherwise it should be present
					// on the Android app's side
					if v := moves.GetByName(action.MoveItem.Name); v != nil {
						m := *v                         // copy values from the library
						m.Delay = action.MoveItem.Delay // copy delay from a user provided variable
						action.MoveItem = &m
					}
				}
			}
		}
	}
	return sessions, nil
}

func collectMoves(dataDir string) ([]*MoveAction, error) {
	query := path.Join(dataDir, "**/*.qianim")
	matches, err := filepath.Glob(query)
	if err != nil {
		return nil, err
	}

	var items = make([]*MoveAction, len(matches))
	for i := range matches {
		// parsing the parent folder as a motion group name
		parts := strings.Split(matches[i], "/")
		parent := parts[len(parts)-2] // TODO: windows error here

		// parsing the basename as a motion name
		basename := parts[len(parts)-1]
		name := strings.Replace(basename, filepath.Ext(basename), "", 1)

		// appending a motion
		items[i] = &MoveAction{
			ID:       uuid.Must(uuid.NewRandom()),
			FilePath: matches[i],
			Group:    parent,
			Name:     name,
		}
	}

	return items, err
}

// Session

// Session represents a session with a child, a set of questions and simple answers which
// are accompanied with moves by a robot to make the conversation a lively one.
type Session struct {
	ID          uuid.UUID      `json:"ID" form:"ID"`
	Name        string         `json:"Name" form:"Name" binding:"required"`
	Description string         `json:"Description" form:"Description"`
	Items       []*SessionItem `json:"Items" form:"Items"`
}

//// Sessions is a wrapper struct around an array of sessions with helpful methods.
//type Sessions []*Session
//
//// GetInstructionByID looks for a top level instruction, which unites Say and Move actions
//// and presents them as a union of two actions, so both actions should be executed.
//func (ss Sessions) GetInstructionByID(id uuid.UUID) *SayAndMoveAction {
//	for _, session := range ss {
//		for _, item := range session.Items {
//			for _, action := range item.Actions {
//				if !action.IsNil() && action.GetID() == id {
//					return action
//				}
//			}
//		}
//	}
//	return nil
//}
//
//func (ss Sessions) GetSessionByID(id string) (*Session, error) {
//	uid, err := uuid.Parse(id)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, s := range ss {
//		if s.ID == uid {
//			return s, nil
//		}
//	}
//
//	return nil, fmt.Errorf("not found")
//}
//
//func (ss Sessions) UpdateSessionByID(updatedSession *Session) error {
//	for _, item := range updatedSession.Items {
//		if (item.ID == uuid.UUID{}) {
//			item.ID = uuid.Must(uuid.NewRandom())
//		}
//
//		for _, action := range item.Actions {
//			if (action.ID == uuid.UUID{}) {
//				action.ID = uuid.Must(uuid.NewRandom())
//			}
//
//			if action.SayItem != nil {
//				if (action.SayItem.ID == uuid.UUID{}) {
//					action.SayItem.ID = uuid.Must(uuid.NewRandom())
//				}
//			}
//
//			if action.MoveItem != nil {
//				if (action.MoveItem.ID == uuid.UUID{}) {
//					action.MoveItem.ID = uuid.Must(uuid.NewRandom())
//				}
//			}
//		}
//	}
//
//	for _, s := range ss {
//		if s.ID == updatedSession.ID {
//			*s = *updatedSession
//		}
//	}
//
//	return nil
//}
//
//func (ss Sessions) DeleteSessionByID(id string) ([]*Session, error) {
//	uid, err := uuid.Parse(id)
//	if err != nil {
//		return nil, err
//	}
//
//	_, err = ss.GetSessionByID(id)
//	if err != nil {
//		return nil, err
//	}
//
//	newSessions := []*Session{}
//
//	for _, s := range ss {
//		if s.ID == uid {
//			continue
//		}
//		newSessions = append(newSessions, s)
//	}
//
//	return newSessions, nil
//}

// SessionItem represents a single unit of a session, it's a question and positive and negative
// answers accompanied with a robot's moves which are represented in the web UI as a set of buttons.
type SessionItem struct {
	ID      uuid.UUID
	Actions []*SayAndMoveAction // the first item of Actions is the main item, usually, it's the main question
	// of the session item, other actions are some kind of conversation supportive answers
}

//type SessionItem struct {
//	ID             uuid.UUID
//	Question       *SayAndMoveAction
//	PositiveAnswer *SayAndMoveAction
//	NegativeAnswer *SayAndMoveAction
//}

type SessionStore struct {
	filepath string
	Sessions []*Session
	mu       sync.RWMutex
}

func NewSessionStore(fpath string) (*SessionStore, error) {
	var file *os.File
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		file, err = os.Create(fpath)
		if err != nil {
			return nil, fmt.Errorf("can't create a session store at %s: %v", fpath, err)
		}
	} else {
		file, err = os.Open(fpath)
	}
	defer file.Close()

	sessions := []*Session{}
	if err = json.NewDecoder(file).Decode(&sessions); err != nil && err != io.EOF {
		return nil, fmt.Errorf("can't decode sessions from %s: %v", fpath, err)
	}

	store := &SessionStore{
		filepath: fpath,
		Sessions: sessions,
	}

	return store, nil
}

// GetInstruction looks for a top level instruction, which unites Say and Move actions
// and presents them as a union of two actions, so both actions should be executed.
func (s *SessionStore) GetInstruction(id uuid.UUID) *SayAndMoveAction {
	for _, session := range s.Sessions {
		for _, item := range session.Items {
			for _, action := range item.Actions {
				if !action.IsNil() && action.GetID() == id {
					return action
				}
			}
		}
	}
	return nil
}

func (s *SessionStore) Get(id string) (*Session, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	for _, s := range s.Sessions {
		if s.ID == uid {
			return s, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (s *SessionStore) Create(newSession *Session) error {
	newSession.ID = uuid.Must(uuid.NewRandom())
	s.Sessions = append(s.Sessions, newSession)
	return s.dump()
}

func (s *SessionStore) Update(updatedSession *Session) error {
	for _, item := range updatedSession.Items {
		if (item.ID == uuid.UUID{}) {
			item.ID = uuid.Must(uuid.NewRandom())
		}

		for _, action := range item.Actions {
			if (action.ID == uuid.UUID{}) {
				action.ID = uuid.Must(uuid.NewRandom())
			}

			if action.SayItem != nil {
				if (action.SayItem.ID == uuid.UUID{}) {
					action.SayItem.ID = uuid.Must(uuid.NewRandom())
				}
			}

			if action.MoveItem != nil {
				if (action.MoveItem.ID == uuid.UUID{}) {
					action.MoveItem.ID = uuid.Must(uuid.NewRandom())
				}
			}
		}
	}

	for _, s := range s.Sessions {
		if s.ID == updatedSession.ID {
			*s = *updatedSession
		}
	}

	return s.dump()
}

func (s *SessionStore) Delete(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	_, err = s.Get(id)
	if err != nil {
		return err
	}

	newSessions := []*Session{}

	for _, s := range s.Sessions {
		if s.ID == uid {
			continue
		}
		newSessions = append(newSessions, s)
	}

	s.Sessions = newSessions

	return s.dump()
}

func (s *SessionStore) dump() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Create(s.filepath)
	if err != nil {
		return fmt.Errorf("failed to open a file: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(s.Sessions)
}

//func (s *SessionStore) load() error {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//
//	f, err := os.Open(s.filepath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	return json.NewDecoder(f).Decode(&s.Sessions)
//}
