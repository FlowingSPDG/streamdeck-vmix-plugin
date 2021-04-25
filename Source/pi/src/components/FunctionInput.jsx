import React from 'react';

// FunctionInput 
class FunctionInput extends React.Component {
    render() {
      return <div className="sdpi-item">
        <div className="sdpi-item-label">Selected</div>
        <input className="sdpi-item-value" readOnly value={this.props.input_key}></input>
      </div>
    }
}

export default FunctionInput