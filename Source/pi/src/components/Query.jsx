import React from 'react';
import Input from './Input'

// Queries
class Queries extends React.Component {
    constructor(props){
      super(props)
    }
    render() {
      return <div>
        <div className="sdpi-item">
          <div className="sdpi-item-label">Query</div>
          <button className="sdpi-item-value" onClick={()=>{this.props.addQuery()}}>Add</button>
        </div>
          { this.props.queries.map((q,index) => 
          <div className="sdpi-item">
            <div className="sdpi-item-label">Query : {index} </div>
              <Query query_key={q.key} query_value={q.value} ></Query>
            </div>
        ) }
      </div>
    }
}
  
class Query extends React.Component {
    render(){
      return <div>
      <Input type="text" className="sdpi-item-value" label="Key" value={this.props.query_key} />
      <Input type="text" className="sdpi-item-value" label="Value" value={this.props.query_value} />
    </div>
    }
}

export { Query, Queries }