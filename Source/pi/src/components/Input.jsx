import React from 'react';

class Input extends React.Component {
    render() {
      return <li id={this.props.id} input_key={this.props.input_key} className={this.props.selected ? "selected" : ""} onClick={this.props.onClick} style={{ textAlign:"left" }} >{this.props.content}</li>
    }
}

export default Input