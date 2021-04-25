import React from 'react';
import FunctionName from './FunctionName';
import {action} from '@storybook/addon-actions';

export default {
  title: 'Function'
};

export const FunctionNames = () => (
  <FunctionName funcName="PreviewInput" onChange={action}/>
);