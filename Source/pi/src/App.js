import React from 'react';
import './App.css';

export class App extends React.Component {
  constructor(props){
    super(props)
    this.state = {
      // vMix variable definitions
      functionName: "PreviewInput", // Function Name
      functionInput: "", // Function Input
      queries: [], // Function query
      inputs: [], // Inputs global variable. This can be button setting
      use_tally_preview: true,
      use_tally_program: true
    }

    this.FunctionNameChange = this.FunctionNameChange.bind(this);
    this.FunctionInputChange = this.FunctionInputChange.bind(this);
    this.tallyPreviewCheckChange = this.tallyPreviewCheckChange.bind(this);
    this.tallyProgramCheckChange = this.tallyProgramCheckChange.bind(this);
  }

  FunctionNameChange(funcName){
    this.setState({functionName:funcName})
  }

  FunctionInputChange(key){
    this.setState({functionInput:key})
  }

  tallyPreviewCheckChange(checked){
    this.setState({use_tally_preview:checked})
  }

  tallyProgramCheckChange(checked){
    this.setState({use_tally_program:checked})
  }

  render(){
    return (
      <div className="App">
      {/* Wrapper starts from here... */}
  
        <div className="sdpi-wrapper">
      
          {/* vMix Function name field(e.g. "Cut"). variable:FunctionName */}
          <FunctionName funcName={this.state.functionName} onChange={this.FunctionNameChange}></FunctionName>
  
          {/* Inputs List. First element should be empty */}
          <InputList inputs={this.state.inputs} selected_key={this.state.functionInput} setSelected={this.FunctionInputChange} ></InputList>
      
          {/* Selected input key(above) */}
          <FunctionInput input_key={this.state.functionInput}></FunctionInput>
      
          {/* Use Tally*/}
          <div type="checkbox" className="sdpi-item">
            <div className="sdpi-item-label">Tally</div>
            <TallyCheck checked={this.state.use_tally_preview} defaultChecked={true} onChange={this.tallyPreviewCheckChange} label="Preview"></TallyCheck>
            <TallyCheck checked={this.state.use_tally_program} defaultChecked={true} onChange={this.tallyProgramCheckChange} label="Program"></TallyCheck>
          </div>
          
      
          {/* Save */}
          <div className="sdpi-item">
            <div className="sdpi-item-label">Save</div>
            <button className="sdpi-item-value">Click to save</button>
          </div>
        </div>
      </div>
    );
  }
}

// FunctionName Function Name such as "PreviewInput".
class FunctionName extends React.Component {
  constructor(props){
    super(props)
    this.state = {funcName: props.funcName ? props.funcName : "Previewinput"}

    this.handleChange = this.handleChange.bind(this);
  }
  handleChange(event){
    this.setState({funcName:event.target.value})
    this.props.onChange(event.target.value)
  }
  render() {
    return <div className="sdpi-item">
      <div className="sdpi-item-label">Function Name</div>
      <input className="sdpi-item-value" value={this.state.funcName} onChange={this.handleChange}></input>
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
    this.state = {
      selected_key: props.selected_key // Selected input key
    }

    this.setSelected = this.setSelected.bind(this);
  }
  setSelected(key){
    this.setState({ selected_key:key })
    this.props.setSelected(key)
  }
  render() {
    return <div type="list" className="sdpi-item list">
      <div className="sdpi-item-label">Inputs List</div>
      <div className="sdpi-item-value single-select" type="">
        <Input onClick={()=>{this.setSelected("")}} selected={this.state.selected_key === "" } input_key={""} id={`input_list_NONE`} content={`NONE`} ></Input>
        <Input onClick={()=>{this.setSelected("0")}} selected={this.state.selected_key === "0" } input_key={"0"} id={`input_list_PRV`} content={`Preview`} ></Input>
        <Input onClick={()=>{this.setSelected("-1")}} selected={this.state.selected_key === "-1" } input_key={"-1"} id={`input_list_PGM`} content={`Program`} ></Input>
        { this.props.inputs.map((input) => <Input onClick={()=>{this.setSelected(input.Key)}} selected={this.state.selected_key === input.Key } input_key={input.key} id={`input_list_${input.Number}`} content={`${input.Number} : ${input.Name}`} ></Input>) }
      </div>
    </div>
  }
}

class Input extends React.Component {
  render() {
    return <li id={this.props.id} key={this.props.input_key} className={this.props.selected ? "selected" : ""} onClick={this.props.onClick} >{this.props.content}</li>
  }
}

// TallyCheck Tally checkbox
class TallyCheck extends React.Component {
  constructor(props){
    super(props)
    this.state = {
      checked: props.checked
    }

    this.handleChange = this.handleChange.bind(this);
  }
  handleChange(event){
    this.setState({checked:!this.state.checked})
    this.props.onChange(this.state.checked)
  }
  render() {
    // Not read-only actually
    return <div type="checkbox" className="sdpi-item">
        <input className="sdpi-item-value" type="checkbox" checked={this.state.checked} readOnly ></input>
        <label><span onClick={this.handleChange}></span>{this.props.label}</label>
      </div>
  }
}

  
export default App;
  