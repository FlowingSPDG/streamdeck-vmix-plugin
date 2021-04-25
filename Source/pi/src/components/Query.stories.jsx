import React from 'react';
import { Query, Queries } from './Query';
import {action} from '@storybook/addon-actions';

export default {
  title: 'Query'
};

export const QueriesStory = () => (
  <div>
    <Queries addQuery={(e)=>{action("Query",e)}} queries={[{"key":"KEY", "value": "VAL"}]}/>
  </div>
);

export const QueryStory = () => (
    <div>
      <Query addQuery={(e)=>{action("Query",e)}} value={"VALUE"}/>
    </div>
);