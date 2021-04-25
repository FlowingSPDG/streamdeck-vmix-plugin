import React from 'react';
import { Query, Queries } from './Query';
import {action} from '@storybook/addon-actions';

export default {
  title: 'Query'
};

export const QueriesStory = () => (
  <div>
    <Queries addQuery={action} queries={[{"key":"KEY", "value": "VAL"}]}/>
  </div>
);

export const QueryStory = () => (
    <div>
      <Query addQuery={action} key="KEY" value="VALUE"/>
    </div>
);