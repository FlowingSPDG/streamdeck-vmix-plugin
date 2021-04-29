package main

import (
	"context"
	"encoding/xml"
	"log"
	"strconv"
	"time"

	"github.com/FlowingSPDG/vmix-go/common/models"
	vmixtcp "github.com/FlowingSPDG/vmix-go/tcp"
)

type tally int

const (
	// Inactive tally status inactive(GREY)
	Inactive tally = iota
	// Preview tally status Preview(GREEN)
	Preview
	// Program tally status Program(RED)
	Program
)

type input struct {
	Name         string `xml:",chardata"`
	Key          string `xml:"key,attr"`
	Number       int    `xml:"number,attr"`
	TallyPreview bool   `xml:"-"`
	TallyProgram bool   `xml:"-"`
}

func vMixGoroutine(ctx context.Context) error {
	// reconnect
	var err error
	vMix, err = vmixtcp.New("localhost")
	if err != nil {
		return err
	}
	defer vMix.Close()

	// re-subscribe
	if err = vMix.SUBSCRIBE(vmixtcp.EVENT_TALLY, ""); err != nil {
		return err
	}

	// We use Tally for checking input added or deleted.
	vMix.Register(vmixtcp.EVENT_TALLY, func(r *vmixtcp.Response) {
		log.Println("TALLY updated. Refreshing... ", r)
		if err := vMix.XML(); err != nil {
			log.Println("Failed to send XMLPATH:", err)
		}
	})

	// Check FUNCTION Command response
	vMix.Register(vmixtcp.EVENT_FUNCTION, func(r *vmixtcp.Response) {
		log.Println("FUNCTION Response received : ", r)
	})

	// If we receive XMLTEXT...
	vMix.Register(vmixtcp.EVENT_XML, func(r *vmixtcp.Response) {
		// log.Println("XML response received:", r)
		x := models.APIXML{}
		if err := xml.Unmarshal([]byte(r.Response), &x); err != nil {
			log.Println("Failed to unmarshal XML:", err)
		}
		newinputs := make([]input, len(x.Inputs.Input))
		for k, v := range x.Inputs.Input {
			num, _ := strconv.Atoi(v.Number)
			newinputs[k] = input{
				Name:         v.Text,
				Key:          v.Key,
				Number:       num,
				TallyPreview: x.Preview == v.Number,
				TallyProgram: x.Active == v.Number,
			}
		}
		settings.Inputs = newinputs
		shouldUpdate = true
	})
	// timeout
	time.Sleep(time.Second)

	// run
	return vMix.Run(ctx)
}
