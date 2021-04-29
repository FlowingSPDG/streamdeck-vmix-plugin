package main

import (
	"context"
	"encoding/xml"
	"log"
	"strconv"
	"strings"
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
	if err = vMix.SUBSCRIBE(vmixtcp.EVENT_ACTS, ""); err != nil {
		return err
	}

	// We use Tally for checking input added or deleted.
	vMix.Register(vmixtcp.EVENT_TALLY, func(r *vmixtcp.Response) {
		log.Println("TALLY updated. Refreshing... ", r)
		if err := vMix.XML(); err != nil {
			log.Println("Failed to send XMLPATH:", err)
		}
	})

	vMix.Register(vmixtcp.EVENT_ACTS, func(r *vmixtcp.Response) {
		log.Printf("ACT updated... %#v\n ", r)
		resps := strings.Split(strings.TrimLeft(r.Response, " "), " ")
		if len(resps) != 3 {
			log.Println("Unknown ACT length", r.Response)
			return
		}
		if resps[0] == "InputPreview" {
			inputNum := resps[1]
			enabled := resps[2] == "1"
			for _, v := range settings.Inputs {
				if enabled {
					v.TallyPreview = strconv.Itoa(v.Number) == inputNum
				}
			}
			shouldUpdate = true
		} else if resps[0] == "Input" {
			inputNum := resps[1]
			enabled := resps[2] == "1"
			for _, v := range settings.Inputs {
				if enabled {
					v.TallyProgram = strconv.Itoa(v.Number) == inputNum
				}
			}
			shouldUpdate = true
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
	})
	// timeout
	time.Sleep(time.Second)

	// run
	return vMix.Run(ctx)
}
