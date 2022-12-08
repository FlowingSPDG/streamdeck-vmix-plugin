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

var (
	inputs = make([]input, 0)
)

type input struct {
	Name         string `json:"name" xml:",chardata"`
	Key          string `json:"key" xml:"key,attr"`
	Number       int    `json:"number" xml:"number,attr"`
	TallyPreview bool   `json:"tally_preview" xml:"-"`
	TallyProgram bool   `json:"tally_program" xml:"-"`
}

func vMixGoroutine(ctx context.Context) error {
	// 何度も再接続したくないので、既に接続が確立していたらやめる
	if vMix != nil {
		return nil
	}

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
		log.Println("XML response received:", r)
		x := models.APIXML{}
		if err := xml.Unmarshal([]byte(r.Response), &x); err != nil {
			log.Println("Failed to unmarshal XML:", err)
		}
		newinputs := make([]input, 0, len(x.Inputs.Input))
		for _, v := range x.Inputs.Input {
			num, _ := strconv.Atoi(v.Number)
			newinputs = append(newinputs, input{
				Name:         v.Text,
				Key:          v.Key,
				Number:       num,
				TallyPreview: x.Preview == v.Number,
				TallyProgram: x.Active == v.Number,
			})
		}
		inputs = newinputs

		shouldUpdate = true
	})
	// timeout
	time.Sleep(time.Second)

	// run
	return vMix.Run(ctx)
}
