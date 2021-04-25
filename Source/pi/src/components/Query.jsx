import React from 'react';
import Form from 'react-bootstrap/Form'
import Col from 'react-bootstrap/Col'
import Row from 'react-bootstrap/Row'


// Queries
class Queries extends React.Component {
    render() {
      return <div>
        <div className="sdpi-item">
          <div className="sdpi-item-label">Query</div>
          <button className="sdpi-item-value" onClick={()=>{this.props.addQuery()}}>Add</button>
        </div>
          { this.props.queries.map((q,index) => 
          <div className="sdpi-item" key={`item_${q.key}_${index}`}>
            <div className="sdpi-item-label" key={`label_${q.key}_${index}`}>Query : {index} </div>
              <Query query_key={q.key} query_value={q.value} key={`query_${q.key}_${index}`}></Query>
            </div>
        ) }
      </div>
    }
}
  
class Query extends React.Component {
    render(){
      return  <Form>
      <Form.Row style={{display: "flex"}}>
        <Col>
        <Form.Control placeholder="Key" style={{minWidth:null}}/>
        </Col>
        <Col>
        <Form.Control placeholder="Duration" />
        </Col>
      </Form.Row>
    </Form>
    }
}

export { Query, Queries }