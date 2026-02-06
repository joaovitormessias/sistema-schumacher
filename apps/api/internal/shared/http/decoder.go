package httpx

import (
  "encoding/json"
  "errors"
  "io"
  "net/http"

  "github.com/go-chi/chi/v5"
  "github.com/google/uuid"
)

var ErrEmptyBody = errors.New("empty body")

func DecodeJSON(r *http.Request, dst interface{}) error {
  if r.Body == nil {
    return ErrEmptyBody
  }
  defer r.Body.Close()

  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()

  if err := dec.Decode(dst); err != nil {
    if errors.Is(err, io.EOF) {
      return ErrEmptyBody
    }
    return err
  }

  // Ensure there is no extra trailing data
  if err := dec.Decode(&struct{}{}); err != io.EOF {
    if err == nil {
      return errors.New("invalid trailing data")
    }
    return err
  }

  return nil
}

func ParseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
  val := chi.URLParam(r, key)
  return uuid.Parse(val)
}
