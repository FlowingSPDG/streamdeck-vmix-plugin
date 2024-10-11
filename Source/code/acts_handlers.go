package stdvmix

import "strconv"

type OnActivatorFunc func(args []string) (on bool)

func newInputAndBoolHandler(activatorName string, input int, on bool) OnActivatorFunc {
	return func(args []string) bool {
		if len(args) != 3 {
			return false
		}
		if args[0] != activatorName {
			return false
		}
		inputNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return false
		}
		if input != inputNumber {
			return false
		}
		isOn := args[2] == "1"
		return isOn == on
	}
}
func newInputAndFloatHandler(activatorName string, input int, f float64) OnActivatorFunc {
	return func(args []string) bool {
		if len(args) != 3 {
			return false
		}
		if args[0] != activatorName {
			return false
		}
		inputNumber, err := strconv.Atoi(args[1])
		if err != nil {
			return false
		}
		if input != inputNumber {
			return false
		}
		value, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return false
		}
		// 本当は大小判定も欲しいかも?
		return f == value
	}
}
func newFloatHandler(activatorName string, f float64) OnActivatorFunc {
	return func(args []string) bool {
		if len(args) != 2 {
			return false
		}
		if args[0] != activatorName {
			return false
		}
		value, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return false
		}
		// 本当は大小判定も欲しいかも?
		return f == value
	}
}
func newBoolHandler(activatorName string, b bool) OnActivatorFunc {
	return func(args []string) bool {
		if len(args) != 2 {
			return false
		}
		if args[0] != activatorName {
			return false
		}
		value := args[1] == "1"
		return value
	}
}

func NewInputHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Input", input, true)
}
func NewInputMix2Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix2", input, true)
}
func NewInputMix3Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix3", input, true)
}
func NewInputMix4Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix4", input, true)
}
func NewInputMix5Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix5", input, true)
}
func NewInputMix6Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix6", input, true)
}
func NewInputMix7Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix7", input, true)
}
func NewInputMix8Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix8", input, true)
}
func NewInputMix9Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix9", input, true)
}
func NewInputMix10Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix10", input, true)
}
func NewInputMix11Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix11", input, true)
}
func NewInputMix12Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix12", input, true)
}
func NewInputMix13Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix13", input, true)
}
func NewInputMix14Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix14", input, true)
}
func NewInputMix15Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix15", input, true)
}
func NewInputMix16Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMix16", input, true)
}
func NewInputPreviewHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreview", input, true)
}
func NewInputPreviewMix2Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix2", input, true)
}
func NewInputPreviewMix3Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix3", input, true)
}
func NewInputPreviewMix4Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix4", input, true)
}
func NewInputPreviewMix5Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix5", input, true)
}
func NewInputPreviewMix6Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix6", input, true)
}
func NewInputPreviewMix7Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix7", input, true)
}
func NewInputPreviewMix8Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix8", input, true)
}
func NewInputPreviewMix9Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix9", input, true)
}
func NewInputPreviewMix10Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix10", input, true)
}
func NewInputPreviewMix11Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix11", input, true)
}
func NewInputPreviewMix12Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix12", input, true)
}
func NewInputPreviewMix13Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix13", input, true)
}
func NewInputPreviewMix14Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix14", input, true)
}
func NewInputPreviewMix15Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix15", input, true)
}
func NewInputPreviewMix16Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPreviewMix16", input, true)
}

func NewInputDynamic1Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputDynamic1", input, true)
}
func NewInputDynamic2Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputDynamic2", input, true)
}
func NewInputDynamic3Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputDynamic3", input, true)
}
func NewInputDynamic4Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputDynamic4", input, true)
}

func NewInputPlayingHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputPlaying", input, true)
}
func NewInputVolumeHandler(input int, volume float64) OnActivatorFunc {
	return newInputAndFloatHandler("InputVolume", input, volume)
}

func NewInputAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputAudio", input, true)
}

func NewInputAudioAutoHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputAudioAuto", input, true)
}
func NewInputAudioSoloHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputSolo", input, true)
}

func NewInputHeadphonesHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputHeadphones", input, true)
}

func NewInputMasterAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputMasterAudio", input, true)
}
func NewinputBusAAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusAAudio", input, true)
}

func NewInputBusBAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusBAudio", input, true)
}
func NewInputBusCAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusCAudio", input, true)
}

func NewInputBusDAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusDAudio", input, true)
}
func NewInputBusEAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusEAudio", input, true)
}
func NewInputBusFAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusFAudio", input, true)
}
func NewInputBusGAudioHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("InputBusGAudio", input, true)
}

// InputVolumeChannelMixer1~16

func NewOverlay1Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay1", input, true)
}
func NewOverlay2Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay2", input, true)
}
func NewOverlay3Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay3", input, true)
}
func NewOverlay4Handler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay4", input, true)
}

func NewOverlay1AnyHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay1Any", input, true)
}
func NewOverlay2AnyHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay2Any", input, true)
}
func NewOverlay3AnyHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay3Any", input, true)
}
func NewOverlay4AnyHandler(input int) OnActivatorFunc {
	return newInputAndBoolHandler("Overlay4Any", input, true)
}

func NewFadeToBlackHandler() OnActivatorFunc {
	return newBoolHandler("FadeToBlack", true)
}
func NewRecordingHandler() OnActivatorFunc {
	return newBoolHandler("Recording", true)
}

func NewStreamingHandler() OnActivatorFunc {
	return newBoolHandler("Streaming", true)
}

func NewExternalHandler() OnActivatorFunc {
	return newBoolHandler("External", true)
}

func NewMulticorderHandler() OnActivatorFunc {
	return newBoolHandler("Multicorder", true)
}

func NewFullscreenHandler() OnActivatorFunc {
	return newBoolHandler("Fullscreen", true)
}

func NewMasterVolumeHandler(volume float64) OnActivatorFunc {
	return newFloatHandler("MasterVolume", volume)
}

// 死ぬほどあるので省略
