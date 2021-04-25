import React from 'react';
import Input from './Input'

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

export default InputList