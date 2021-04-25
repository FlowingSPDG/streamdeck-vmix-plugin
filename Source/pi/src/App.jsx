import React from 'react';
import { hot } from 'react-hot-loader';
import './App.css';

// components
import FunctionInput from './components/FunctionInput'
import FunctionName from './components/FunctionName'
import InputList from './components/InputList'
import { Queries } from './components/Query'
import TallyCheck from './components/TallyCheck'

export class App extends React.Component {
  constructor(props){
    // super
    super(props)

    // States
    this.state = {
      // vMix variable definitions
      functionName: "PreviewInput", // Function Name
      functionInput: "", // Function Input
      queries: [], // Function query
      inputs: [], // Inputs global variable. This can be button setting
      use_tally_preview: true,
      use_tally_program: true
    }

    // Bind "this"(fuck.)
    this.FunctionNameChange = this.FunctionNameChange.bind(this);
    this.FunctionInputChange = this.FunctionInputChange.bind(this);
    this.tallyPreviewCheckChange = this.tallyPreviewCheckChange.bind(this);
    this.tallyProgramCheckChange = this.tallyProgramCheckChange.bind(this);
    this.addQuery = this.addQuery.bind(this);
    this.deleteQuery = this.deleteQuery.bind(this);
    this.handleKeyChange = this.handleKeyChange.bind(this);
    this.handleValueChange = this.handleValueChange.bind(this);

    this.pluginAction = null
    this.uuid = ''
    this.context = ""
  }

  componentDidMount(){
      if (window.$SD) {
        window.$SD.on('connected', (jsonObj)=> {
          console.log("connected", jsonObj)
          this.uuid = jsonObj['uuid'];
          if (jsonObj.hasOwnProperty('actionInfo')) {
            this.pluginAction = jsonObj.actionInfo['action'];
            this.context = jsonObj.actionInfo['context'];

            if (jsonObj.actionInfo.payload.hasOwnProperty("settings")){
              // Input
              if(jsonObj.actionInfo.payload.settings.functionInput !== "") {
                console.log("Updating functionInput:",jsonObj.actionInfo.payload.settings.functionInput)
                this.setState({functionInput:jsonObj.actionInfo.payload.settings.functionInput})
              }
              if(jsonObj.actionInfo.payload.settings.hasOwnProperty("use_tally_preview")) {
                console.log("Updating use_tally_preview:",jsonObj.actionInfo.payload.settings.use_tally_preview)
                this.setState({use_tally_preview : jsonObj.actionInfo.payload.settings.use_tally_preview})
              }
              if(jsonObj.actionInfo.payload.settings.hasOwnProperty("use_tally_program")) {
                console.log("Updating use_tally_program:",jsonObj.actionInfo.payload.settings.use_tally_program)
                this.setState({use_tally_program : jsonObj.actionInfo.payload.settings.use_tally_program})
              }

              // functionName
              if(typeof(jsonObj.actionInfo.payload.settings.functionName) === "string" && jsonObj.actionInfo.payload.settings.functionName !== ''){
                console.log("Updating functionName:",jsonObj.actionInfo.payload.settings.functionName)
                this.setState({functionName:jsonObj.actionInfo.payload.settings.functionName})
              }

              // Inputs
              if(Array.isArray(jsonObj.actionInfo.payload.settings.inputs)){
                console.log("Updating inputs:", jsonObj.actionInfo.payload.settings.inputs)
                this.setState({inputs:jsonObj.actionInfo.payload.settings.inputs})
              }

              if(Array.isArray(jsonObj.actionInfo.payload.settings.queries)){
                this.setState({queries:jsonObj.actionInfo.payload.settings.queries})
              }
            }
          }
          console.log("current state:",this.state)
          console.log("Requesting force update")
          this.forceUpdate()
          console.log("Force update done")
        });

        window.$SD.on("sendToPropertyInspector", (jsonObj) => {
          console.log("sendToPropertyInspector", jsonObj)
          if(!jsonObj.payload){
            return
          }
          if(jsonObj.event === "sendToPropertyInspector"){
            if(Array.isArray(jsonObj.payload.inputs)){
              if(this.state.inputs.length !== jsonObj.payload.inputs.length){
                this.setState({inputs:jsonObj.payload.inputs})
              }
            }
          }
          else if(jsonObj.event === "didReceiveSettings"){
            console.log("didReceiveSettings", jsonObj.payload)
          }
        })
        window.$SD.on("didReceiveGlobalSettings", (jsonObj) => {
          console.log("didReceiveGlobalSettings")
        })
      }
    }

  saveSettings(){
    console.log("Saving setting")
    if (window.$SD && window.$SD.connection){
      window.$SD.api.sendToPlugin(this.uuid, this.pluginAction,{
        "functionInput":this.state.functionInput,
        "functionName": this.state.functionName,
        "queries":this.state.queries,
        "use_tally_preview":this.state.use_tally_preview,
        "use_tally_program":this.state.use_tally_program,
      })
    }
  }

  FunctionNameChange(funcName){
    this.setState({functionName:funcName}, ()=>{
      this.saveSettings()
    })
  }

  FunctionInputChange(key){
    this.setState({functionInput:key},()=>{
      this.saveSettings()
    })
  }

  tallyPreviewCheckChange(checked){
    this.setState({use_tally_preview: checked}, ()=>{
      this.saveSettings()
    })
  }

  tallyProgramCheckChange(checked){
    this.setState({use_tally_program: checked}, () => {
      this.saveSettings()
    })
  }

  addQuery(){
    const newq = this.state.queries.slice()
    newq.push({key:"Duration",value:"500"})
    this.setState({queries:newq})
  }

  deleteQuery(index){
    const newq = this.state.queries.filter((v,i) => {
      return index !== i
    })
    this.setState({queries:newq})
  }

  handleKeyChange(index, key){
    const newq = this.state.queries.slice()
    newq[index].key = key
    this.setState({queries:newq})
  }

  handleValueChange(index, value){
    const newq = this.state.queries.slice()
    newq[index].value = value
    this.setState({queries:newq})
  }

  render(){
    return (
      <div className="App">
      {/* Wrapper starts from here... */}
        <div className="sdpi-wrapper">
      
          {/* vMix Function name field(e.g. "Cut"). variable:FunctionName */}
          <FunctionName funcName={this.state.functionName} funcNameChange={this.FunctionNameChange} />
  
          {/* Inputs List. First element should be empty */}
          <InputList inputs={this.state.inputs} selected_key={this.state.functionInput} setSelected={this.FunctionInputChange} />
      
          {/* Selected input key(above) */}
          <FunctionInput input_key={this.state.functionInput} />
      
          {/* Use Tally */}
          <div type="checkbox" className="sdpi-item">
            <div className="sdpi-item-label">Tally</div>
            <TallyCheck checked={this.state.use_tally_preview} defaultChecked={true} onChange={this.tallyPreviewCheckChange} label="Preview" />
            <TallyCheck checked={this.state.use_tally_program} defaultChecked={true} onChange={this.tallyProgramCheckChange} label="Program" />
          </div>

          {/* Function query */}
          <Queries QueryKeyChange={this.handleKeyChange} QueryValueChange={this.handleValueChange} queries={this.state.queries} addQuery={this.addQuery} deleteQuery={this.deleteQuery} />
        </div>
      </div>
    );
  }
}

export default hot(module)(App);
