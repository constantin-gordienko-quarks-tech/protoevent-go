package cloudevent

import (
	"time"

	"github.com/streadway/amqp"

	"github.com/quarks-tech/protoevent-go/pkg/event"
)

type Formatter struct{}

func (Formatter) Format(meta *event.Metadata, data []byte) amqp.Publishing {
	return amqp.Publishing{
		Type:        meta.Type,
		ContentType: meta.DataContentType,
		Headers:     buildPublishingHeaders(meta),
		Body:        data,
	}
}

func buildPublishingHeaders(meta *event.Metadata) amqp.Table {
	return amqp.Table{
		"cloudEvents:specversion": meta.SpecVersion,
		"cloudEvents:time":        meta.Time.Format(time.RFC3339),
		"cloudEvents:id":          meta.ID,
		"cloudEvents:source":      meta.Source,
		"cloudEvents:subject":     meta.Subject,
	}
}
