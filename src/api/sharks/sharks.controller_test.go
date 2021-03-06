package sharks

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sharkies/db"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ----------------------------------------------------------------------

var (
	err        error
	dbUser     string
	dbPassword string
	dbHost     string
	dbPort     string
	dbName     string
	e          *echo.Echo
)

func init() {
	// DB Config
	dbUser = os.Getenv("TEST_DB_USER")
	dbPassword = os.Getenv("TEST_DB_PASSWORD")
	dbHost = os.Getenv("TEST_DB_HOST")
	dbPort = os.Getenv("TEST_DB_PORT")
	dbName = os.Getenv("TEST_DB_NAME")

	// DB
	db.Setup(dbUser, dbPassword, dbHost, dbPort, dbName)

	// Echo
	e = echo.New()
}

// ----------------------------------------------------------------------

/*
	!: Add new tests here
*/

func Test_getSharks(t *testing.T) {

	t.Run("returns the list of sharks at index", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/")

		expectedCount := 2
		var got []Shark

		// Assertions
		if assert.NoError(t, getSharks(c)) {

			assert.Equal(t, http.StatusOK, rec.Code)

			if err := json.Unmarshal([]byte(rec.Body.String()), &got); err != nil {
				log.Fatal("error parsing response body -> ", err)
			}

			assert.Equal(t, expectedCount, len(got))
		}
	})

	t.Run("returns the list of sharks at /sharks", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks")

		expectedCount := 2
		var got []Shark

		// Assertions
		if assert.NoError(t, getSharks(c)) {

			assert.Equal(t, http.StatusOK, rec.Code)

			if err := json.Unmarshal([]byte(rec.Body.String()), &got); err != nil {
				log.Fatal("error parsing response body -> ", err)
			}

			assert.Equal(t, expectedCount, len(got))
		}
	})

	t.Run("returns an empty slice when non-existent sharks", func(t *testing.T) {
		shark := new(Shark)
		shark.deleteAll()

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks")

		expectedCount := 0
		var got []Shark

		// Assertions
		if assert.NoError(t, getSharks(c)) {
			if assert.Equal(t, http.StatusOK, rec.Code) {

				if err := json.Unmarshal([]byte(rec.Body.String()), &got); err != nil {
					log.Fatal("error parsing response body -> ", err)
				}

				assert.Equal(t, expectedCount, len(got))

				// adds the shark and go back to the previous state
				newshark1 := &Shark{Name: "Basking Shark", Bname: "Cetorhinus maximus", Description: "Description of basking shark", Image: "Image of basking shark"}
				newshark2 := &Shark{Name: "Zebra Bullhead Shark", Bname: "Heterodontus zebra", Description: "Description of zebra shark", Image: "Image of zebra shark"}
				shark.addShark(newshark1)
				shark.addShark(newshark2)
			}
		}
	})
}

func Test_getShark(t *testing.T) {

	t.Run("returns a shark with given id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")

		s := new(Shark)
		sharks, _ := s.findAll()

		rand.Seed(time.Now().UnixNano())
		idx := 0 + rand.Intn(len(sharks)-0+1-1)

		genID := sharks[idx].ID
		id := strconv.FormatInt(genID, 10)

		c.SetParamValues(id)

		// Assertions
		if assert.NoError(t, getShark(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			assert.Contains(t, rec.Body.String(), id)
		}
	})

	t.Run("returns an error when invalid id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("a")

		expected := map[string]string{"error": "error parsing id"}

		// Assertions
		if assert.NoError(t, getShark(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			expectedJSON, err := json.Marshal(expected)

			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}

			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})

	t.Run("returns an error when non-existent id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("-1")

		expected := map[string]string{"error": "error obtaining shark"}

		// Assertions
		if assert.NoError(t, getShark(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)

			expectedJSON, err := json.Marshal(expected)

			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}

			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})
}

func Test_addShark(t *testing.T) {

	t.Run("returns shark id", func(t *testing.T) {
		newShark := `{"name":"Test Name three","bname":"Test Bname three","description":"Test Description three","image":"Test Image three"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(newShark))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/sharks")

		// Assertions
		if assert.NoError(t, addShark(c)) {

			if assert.Equal(t, http.StatusCreated, rec.Code) {

				var gID map[string]int64
				err := json.Unmarshal([]byte(rec.Body.String()), &gID)
				if err != nil {
					log.Fatal("error parsing response body -> ", err)
				}

				assert.Contains(t, rec.Body.String(), "id")

				// delete the shark and go back to the previous state
				shark := new(Shark)
				id := gID["id"]
				shark.deleteShark(id)
			}
		}
	})
}

func Test_deleteShark(t *testing.T) {

	t.Run("returns shark deleted", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")

		s := new(Shark)
		sharks, _ := s.findAll()

		rand.Seed(time.Now().UnixNano())
		idx := 0 + rand.Intn(len(sharks)-0+1-1)

		genID := sharks[idx].ID
		id := strconv.FormatInt(genID, 10)

		shark := new(Shark)
		currentShark, _ := shark.findByID(genID)

		c.SetParamValues(id)

		sharkJSON := map[string]string{"msg": "shark deleted"}

		// Assertions
		if assert.NoError(t, deleteShark(c)) {
			if assert.Equal(t, http.StatusOK, rec.Code) {
				expectedJSON, err := json.Marshal(sharkJSON)

				if err != nil {
					log.Fatal("error parsing expected response -> ", err)
				}

				assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))

				// adds the shark and go back to the previous state
				shark.addShark(&currentShark)
			}
		}
	})

	t.Run("returns an error when invalid id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("a")

		sharkJSON := map[string]string{"error": "error parsing id"}

		// Assertions
		if assert.NoError(t, deleteShark(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			expectedJSON, err := json.Marshal(sharkJSON)

			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}

			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})

	t.Run("returns an error when non-existent id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("-1")

		sharkJSON := map[string]string{"error": "error deleting shark"}

		// Assertions
		if assert.NoError(t, deleteShark(c)) {
			assert.Equal(t, http.StatusNotFound, rec.Code)
			expectedJSON, err := json.Marshal(sharkJSON)

			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}

			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})
}

func Test_deleteSharks(t *testing.T) {

	t.Run("returns all sharks deleted", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		c := e.NewContext(req, rec)
		c.SetPath("/sharks")

		sharksJSON := map[string]string{"msg": "all sharks deleted"}
		expectedJSON, err := json.Marshal(sharksJSON)
		if err != nil {
			log.Fatal("error parsing response body -> ", err)
		}

		// Assertions
		if assert.NoError(t, deleteSharks(c)) {
			if assert.Equal(t, http.StatusOK, rec.Code) {
				assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))

				// adds the shark and go back to the previous state
				newshark1 := &Shark{Name: "Basking Shark", Bname: "Cetorhinus maximus", Description: "Description of basking shark", Image: "Image of basking shark"}
				newshark2 := &Shark{Name: "Zebra Bullhead Shark", Bname: "Heterodontus zebra", Description: "Description of zebra shark", Image: "Image of zebra shark"}
				shark := new(Shark)
				shark.addShark(newshark1)
				shark.addShark(newshark2)
			}
		}
	})
}

func Test_PatchShark(t *testing.T) {

	t.Run("returns shark patched", func(t *testing.T) {
		newName := `{"name":"Test Name patched two"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(newName))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")

		s := new(Shark)
		sharks, _ := s.findAll()

		rand.Seed(time.Now().UnixNano())
		idx := 0 + rand.Intn(len(sharks)-0+1-1)

		genID := sharks[idx].ID
		log.Printf("%v", sharks[idx])
		id := strconv.FormatInt(genID, 10)

		shark := new(Shark)
		currentShark, _ := shark.findByID(genID)

		c.SetParamValues(id)

		sharkJSON := map[string]string{"msg": "shark patched"}

		// Assertions
		if assert.NoError(t, patchShark(c)) {
			if assert.Equal(t, http.StatusOK, rec.Code) {
				expectedJSON, err := json.Marshal(sharkJSON)

				if err != nil {
					log.Fatal("error parsing expected response -> ", err)
				}

				assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))

				// re-insert the original shark and go back to the previous state
				shark.deleteShark(genID)
				shark.addShark(&currentShark)
			}
		}
	})

	t.Run("returns an error when invalid id", func(t *testing.T) {
		newName := `{"name":"Test Name patched two"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(newName))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("a")

		sharkJSON := map[string]string{"error": "error parsing id"}

		// Assertions
		if assert.NoError(t, patchShark(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			expectedJSON, err := json.Marshal(sharkJSON)
			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}
			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})

	t.Run("returns an error when non-existent id", func(t *testing.T) {
		newName := `{"name":"Test Name patched two"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(newName))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")
		c.SetParamValues("-1")

		sharkJSON := map[string]string{"error": "error patching shark"}

		// Assertions
		if assert.NoError(t, patchShark(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			expectedJSON, err := json.Marshal(sharkJSON)
			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}
			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})

	t.Run("returns an error when modifying id", func(t *testing.T) {
		newID := `{"id":999}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(newID))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/sharks/:id")
		c.SetParamNames("id")

		s := new(Shark)
		sharks, _ := s.findAll()

		rand.Seed(time.Now().UnixNano())
		idx := 0 + rand.Intn(len(sharks)-0+1-1)

		genID := sharks[idx].ID
		id := strconv.FormatInt(genID, 10)

		c.SetParamValues(id)

		sharkJSON := map[string]string{"error": "error patching shark"}

		// Assertions
		if assert.NoError(t, patchShark(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			expectedJSON, err := json.Marshal(sharkJSON)

			if err != nil {
				log.Fatal("error parsing expected response -> ", err)
			}

			assert.Equal(t, string(expectedJSON), strings.TrimSuffix(rec.Body.String(), "\n"))
		}
	})
}
