import React from 'react';
import { hot } from 'react-hot-loader';
import TextField from '@material-ui/core/TextField';
import './App.css';

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
    this.setState({functionName:funcName})
    this.saveSettings()
  }

  FunctionInputChange(key){
    this.setState({functionInput:key})
    this.saveSettings()
  }

  tallyPreviewCheckChange(checked){
    this.setState({use_tally_preview:checked})
    this.saveSettings()
  }

  tallyProgramCheckChange(checked){
    this.setState({use_tally_program:checked})
    this.saveSettings()
  }

  addQuery(){
    this.state.queries.push({key:"",value:""})
    const newq = this.state.queries
    this.setState({queries:newq})
  }

  render(){
    return (
      <div className="App">
      {/* Wrapper starts from here... */}
        <h2>React</h2>
        <div className="sdpi-wrapper">
      
          {/* vMix Function name field(e.g. "Cut"). variable:FunctionName */}
          <FunctionName funcName={this.state.functionName} funcNameChange={this.FunctionNameChange}></FunctionName>
  
          {/* Inputs List. First element should be empty */}
          <InputList inputs={this.state.inputs} selected_key={this.state.functionInput} setSelected={this.FunctionInputChange} ></InputList>
      
          {/* Selected input key(above) */}
          <FunctionInput input_key={this.state.functionInput}></FunctionInput>
      
          {/* Use Tally */}
          <div type="checkbox" className="sdpi-item">
            <div className="sdpi-item-label">Tally</div>
            <TallyCheck checked={this.state.use_tally_preview} defaultChecked={true} onChange={this.tallyPreviewCheckChange} label="Preview"></TallyCheck>
            <TallyCheck checked={this.state.use_tally_program} defaultChecked={true} onChange={this.tallyProgramCheckChange} label="Program"></TallyCheck>
          </div>

          {/* Function query */}
          <Queries queries={this.state.queries} addQuery={this.addQuery} ></Queries>
        </div>
      </div>
    );
  }
}

// FunctionName Function Name such as "PreviewInput".
class FunctionName extends React.Component {
  handleChange = (event) =>{
    this.props.funcNameChange(event.target.value) // Trigger callback
  }
  render() {
    return <div className="sdpi-item">
      <div className="sdpi-item-label">Function Name</div>
      <input className="sdpi-item-value" value={this.props.funcName} onChange={this.handleChange}></input>
    </div>
  }
}

// FunctionInput 
class FunctionInput extends React.Component {
  render() {
    return <div className="sdpi-item">
      <div className="sdpi-item-label">Selected</div>
      <input className="sdpi-item-value" readOnly value={this.props.input_key}></input>
    </div>
  }
}

class InputList extends React.Component {
  constructor(props){
    super(props)
    this.setSelected = this.setSelected.bind(this);
  }
  setSelected(input_key){
    this.props.setSelected(input_key)
  }
  render() {
    return <div type="list" className="sdpi-item list">
      <div className="sdpi-item-label">Inputs List</div>
      <div className="sdpi-item-value single-select" type="">
        <Input onClick={()=>{this.setSelected("")}} selected={this.props.selected_key === "" } input_key={"NONE"} key={"NONE"} id={`input_list_NONE`} content={`NONE`} ></Input>
        <Input onClick={()=>{this.setSelected("0")}} selected={this.props.selected_key === "0" } input_key={"0"} key={"PRV"} id={`input_list_PRV`} content={`Preview`} ></Input>
        <Input onClick={()=>{this.setSelected("-1")}} selected={this.props.selected_key === "-1" } input_key={"-1"} key={"PGM"} id={`input_list_PGM`} content={`Program`} ></Input>
        { this.props.inputs.map((input) => <Input onClick={()=>{this.setSelected(input.Key)}} selected={this.props.selected_key === input.Key } input_key={input.key} key={input.key} id={`input_list_${input.Number}`} content={`${input.Number} : ${input.Name}`} ></Input>) }
      </div>
    </div>
  }
}

class Input extends React.Component {
  render() {
    return <li id={this.props.id} input_key={this.props.input_key} className={this.props.selected ? "selected" : ""} onClick={this.props.onClick} >{this.props.content}</li>
  }
}

// TallyCheck Tally checkbox
class TallyCheck extends React.Component {
  constructor(props){
    super(props)
    this.handleChange = this.handleChange.bind(this);
  }
  handleChange(event){
    // Reverse check
    this.props.onChange(!this.props.checked)
  }
  render() {
    // Not read-only actually
    return <div type="checkbox" className="sdpi-item">
        <input className="sdpi-item-value" type="checkbox" checked={this.props.checked} readOnly ></input>
        <label><span onClick={this.handleChange}></span>{this.props.label}</label>
      </div>
  }
}

// Queries
class Queries extends React.Component {
  constructor(props){
    super(props)
  }
  render() {
    return <div>
      <div className="sdpi-item">
        <div className="sdpi-item-label">Query</div>
        <button className="sdpi-item-value" onClick={()=>{this.props.addQuery()}}>Add</button>
      </div>
        { this.props.queries.map((q,index) => 
        <div className="sdpi-item">
          <div className="sdpi-item-label">Query : {index} </div>
            <Query query_key={q.key} query_value={q.value} ></Query>
          </div>
      ) }
    </div>
  }
}

class Query extends React.Component {
  render(){
    return <div>
    <Input type="text" className="sdpi-item-value" label="Key" value={this.props.query_key} />
    <Input type="text" className="sdpi-item-value" label="Value" value={this.props.query_value} />
  </div>
  }
}

export default hot(module)(App);
