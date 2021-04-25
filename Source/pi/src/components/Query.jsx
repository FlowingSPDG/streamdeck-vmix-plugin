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
    this.props.QueryKeyChange(index,val)
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
          { this.props.queries.map((q,index) => 
        <div className="sdpi-item" key={`item_${q.key}_${index}`}>
          <div className="sdpi-item-label" key={`label_${q.key}_${index}`}>Query : {index} </div>
            <Query handleKeyChange={(val)=>{this.handleKeyChange(index,val)}} handleValueChange={(val)=>{this.handleValueChange(index,val)}}query_key={q.key} query_value={q.value} key={`query_${q.key}_${index}`}></Query>
          </div>
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
  handleKeyChange(val){
    this.props.QueryKeyChange(val)
  }
  handleValueChange(val){
    this.props.QueryValueChange(val)
  }
    render(){
      return  <Form>
      <Form.Row style={{display: "flex"}}>
        <Col>
        <Form.Control placeholder="Key" onChange={this.handleKeyChange} style={{minWidth:null}}/>
        </Col>
        <Col>
        <Form.Control placeholder="Value" onChange={this.handleValueChange} style={{minWidth:null}}/>
        </Col>
      </Form.Row>
    </Form>
    }
}

export { Query, Queries }