import React from 'react';
import Input from './Input';
import {action} from '@storybook/addon-actions';

export default {
  title: 'Input'
};

export const Inputs = () => (
  <div>
    <Input id="hogehoge_ID" input_key="hogehoge_KEY" selected={true} onClick={action} content={"Content"}/>
    <Input id="hogehoge_ID" input_key="hogehoge_KEY" selected={false} onClick={action} content={"Content"}/>
  </div>
);