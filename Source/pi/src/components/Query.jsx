import React from 'react';
import Form from 'react-bootstrap/Form'
import Col from 'react-bootstrap/Col'

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
          <button className="sdpi-item-value" onClick={()=>{this.props.addQuery()}}>Add</button>
        </div>
          { this.props.queries.map((q,index) => {
            return <div className="sdpi-item" key={`item_${index}`}>
            <div className="sdpi-item-label" key={`label_${index}`}>Query : {index} </div>
              <Query QueryKeyChange={(val)=>{this.handleKeyChange(index,val)}} QueryValueChange={(val)=>{this.handleValueChange(index,val)}} query_key={q.key} query_value={q.value} key={`query_${index}`}></Query>
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
  handleKeyChange(e){
    this.props.QueryKeyChange(e.target.value)
  }
  handleValueChange(e){
    this.props.QueryValueChange(e.target.value)
  }
    render(){
      return  <Form>
      <Form.Row style={{display: "flex"}}>
        <Col>
        <Form.Control placeholder="Key" onChange={this.handleKeyChange} style={{minWidth:null}} value={this.props.query_key}/>
        </Col>
        <Col>
        <Form.Control placeholder="Value" onChange={this.handleValueChange} style={{minWidth:null}} value={this.props.query_value} />
        </Col>
      </Form.Row>
    </Form>
    }
}

export { Query, Queries }