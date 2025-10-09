package serialization

import (
	"encoding/json"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
)

// Serializer handles data serialization optimization (Golden Rule #10: Data Serialization)
type Serializer struct {
	useMessagePack bool
}

// NewSerializer creates a new serializer instance
func NewSerializer(useMessagePack bool) *Serializer {
	return &Serializer{
		useMessagePack: useMessagePack,
	}
}

// SerializeResponse serializes data based on client preference
func (s *Serializer) SerializeResponse(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// Check if client accepts MessagePack
	accept := r.Header.Get("Accept")
	useMsgPack := s.useMessagePack && (accept == "application/msgpack" || accept == "application/x-msgpack")

	if useMsgPack {
		// Use MessagePack for faster serialization
		w.Header().Set("Content-Type", "application/msgpack")
		return msgpack.NewEncoder(w).Encode(data)
	}

	// Default to JSON
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// DeserializeRequest deserializes request data based on content type
func (s *Serializer) DeserializeRequest(r *http.Request, dest interface{}) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/msgpack" || contentType == "application/x-msgpack" {
		// Use MessagePack for faster deserialization
		return msgpack.NewDecoder(r.Body).Decode(dest)
	}

	// Default to JSON
	return json.NewDecoder(r.Body).Decode(dest)
}

// GetSupportedContentTypes returns supported content types
func (s *Serializer) GetSupportedContentTypes() []string {
	if s.useMessagePack {
		return []string{"application/json", "application/msgpack"}
	}
	return []string{"application/json"}
}
