import React from 'react';
import InputList from './InputList';
import {action} from '@storybook/addon-actions';

export default {
  title: 'Input'
};

export const InputLists = () => (
    <InputList setSelected={action} selected_key="0" inputs={[{"key":"KEY","Number":0,"Name":"hogehoge"}]}/>
);