import React from 'react';

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

export default TallyCheck