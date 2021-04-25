import React from 'react';
import Form from 'react-bootstrap/Form'

// Queries
class Queries extends React.Component {
  constructor(props){
    super(props)

    this.handleKeyChange = this.handleKeyChange.bind(this);
    this.handleValueChange = this.handleValueChange.bind(this);
  }
  handleKeyChange(index, val){
    this.props.QueryKeyChange(index, val)
  }
  handleValueChange(index, val){
    this.props.QueryValueChange(index, val)
  }

  render() {
      return <div>
        <div className="sdpi-item">
          <div className="sdpi-item-label">Query</div>
          <button className="sdpi-item-value" type="button" onClick={()=>{this.props.addQuery()}}>Add</button>
        </div>
          { this.props.queries.map((q,index) => {
            return <div className="sdpi-item" key={`item_${index}`}>
            <div className="sdpi-item-label" key={`label_${index}`}>Query : {index} </div>
            <Query QueryKeyChange={(val)=>{this.handleKeyChange(index,val)}} QueryValueChange={(val)=>{this.handleValueChange(index,val)}} query_key={q.key} query_value={q.value} key={`query_${index}`} index={index} deleteQuery={()=>{this.props.deleteQuery(index)}}></Query>
            </div>
          }
      ) }
    </div>
  }
}

class Query extends React.Component {
  constructor(props){
    super(props)

    this.handleKeyChange = this.handleKeyChange.bind(this);
    this.handleValueChange = this.handleValueChange.bind(this);
  }
  deleteQuery(index){
    this.props.deleteQuery(index)
  }
  handleKeyChange(e){
    this.props.QueryKeyChange(e.target.value)
  }
  handleValueChange(e){
    this.props.QueryValueChange(e.target.value)
  }
    render(){
      return  <Form>
      <Form.Row style={{display: "flex", margin:"2%", textAlign:"center"}}>
          <Form.Control placeholder="Key" onChange={this.handleKeyChange} style={{width:"25%",marginRight:"3%"}} value={this.props.query_key}/>
          <Form.Control placeholder="Value" onChange={this.handleValueChange} style={{width:"25%",marginRight:"3%"}} value={this.props.query_value} />
          <button className="sdpi-item-value" type="button" onClick={(e)=>{this.deleteQuery(this.props.index)}} style={{width:"23%",marginRight:"3%"}}>Delete</button>
      </Form.Row>
    </Form>
    }
}

export { Query, Queries }