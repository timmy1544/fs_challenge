package api

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/yourusername/task-manager/internal/websocket"
)

// ComputeDelta computes the difference between two structs and returns only changed fields
func ComputeDelta(old, new interface{}) (map[string]interface{}, error) {
	delta := make(map[string]interface{})

	oldVal := reflect.ValueOf(old)
	newVal := reflect.ValueOf(new)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	if oldVal.Type() != newVal.Type() {
		return nil, nil
	}

	oldType := oldVal.Type()
	for i := 0; i < oldVal.NumField(); i++ {
		field := oldType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty and other options
		if idx := 0; idx < len(jsonTag); idx++ {
			if jsonTag[idx] == ',' {
				jsonTag = jsonTag[:idx]
				break
			}
		}

		oldField := oldVal.Field(i)
		newField := newVal.Field(i)

		if !reflect.DeepEqual(oldField.Interface(), newField.Interface()) {
			delta[jsonTag] = newField.Interface()
		}
	}

	return delta, nil
}

// CreateDeltaMessage creates a WebSocket message with delta updates
func CreateDeltaMessage(msgType string, projectID uuid.UUID, taskID *uuid.UUID, oldData, newData interface{}) (*websocket.Message, error) {
	delta, err := ComputeDelta(oldData, newData)
	if err != nil {
		return nil, err
	}

	message := &websocket.Message{
		Type:      msgType,
		ProjectID: projectID,
		TaskID:    taskID,
		Delta:     delta,
	}

	// For creates, include full data
	if msgType == "TASK_CREATED" || msgType == "COMMENT_CREATED" {
		message.FullData = newData
	}

	return message, nil
}
