import React from 'react';

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

export default FunctionName