// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

import {describe, expect, test, vi} from 'vitest';
import {mount} from '@vue/test-utils';
import ActionJobStepList from './ActionJobStepList.vue';
import ActionJobStep from './ActionJobStep.vue';

vi.mock('../utils/time.js', () => ({
  formatDatetime: vi.fn((date) => date.toISOString()),
}));

describe('ActionJobStepList', () => {
  const mockIsExpandable = vi.fn(() => true);
  const mockIsDone = vi.fn(() => true);

  const defaultProps = {
    steps: [
      {status: 'success', summary: 'step 1 - go grocery shopping', duration: '3601s'},
      {status: 'running', summary: 'step 2 - cook food', duration: '301s'},
      {status: 'waiting', summary: 'step 3 - eat food', duration: ''},
    ],
    stepStates: [
      {cursor: null, expanded: false},
      {cursor: null, expanded: false},
      {cursor: null, expanded: false},
    ],
    runStatus: 'running',
    isExpandable: mockIsExpandable,
    isDone: mockIsDone,
    timeVisibleTimestamp: false,
    timeVisibleSeconds: false,
  };

  function createWrapper(props = {}) {
    return mount(ActionJobStepList, {
      props: {
        ...defaultProps,
        ...props,
      },
    });
  }

  test('pass-through to ActionJobStep', () => {
    // ActionJobStepList's tests don't need to validate the functionality of ActionJobStep -- that is the responsibility
    // of its own tests.  But we should validate that ActionJobStepList invokes the child element and passes in relevant
    // data.
    const wrapper = createWrapper();
    const jobSteps = wrapper.findAllComponents(ActionJobStep);
    expect(jobSteps).toHaveLength(3);
    expect(jobSteps[0].props()).toEqual({
      runStatus: 'running',
      isExpandable: mockIsExpandable,
      isDone: mockIsDone,
      stepId: 0,
      status: 'success',
      summary: 'step 1 - go grocery shopping',
      duration: '3601s',
      expanded: false,
      cursor: null,
      timeVisibleTimestamp: false,
      timeVisibleSeconds: false,
    });
  });

  test('render list elements', () => {
    const wrapper = createWrapper();
    expect(wrapper.findAll('.job-step-container')).toHaveLength(1);
    expect(wrapper.findAll('.job-step-section')).toHaveLength(3);
  });

  test('render empty', () => {
    const wrapper = createWrapper({steps: []});
    expect(wrapper.findAll('.job-step-container')).toHaveLength(0);
    expect(wrapper.findAll('.job-step-section')).toHaveLength(0);
  });

  test('pass-through appendLogs', () => {
    const wrapper = createWrapper({
      stepStates: [
        {cursor: 0, expanded: true},
        {cursor: null, expanded: false},
        {cursor: null, expanded: false},
      ],
    });

    expect(wrapper.findAll('.job-log-line').length).toEqual(0);

    const logLines = [
      {index: 1, timestamp: 1765163618, message: 'Starting build'},
      {index: 2, timestamp: 1765163619, message: 'Running tests'},
      {index: 3, timestamp: 1765163620, message: 'Build complete'},
    ];
    wrapper.vm.appendLogs(0, logLines, 1765163618);

    expect(wrapper.findAll('.job-log-line').length).toEqual(3);
  });

  test('toggle visibility of timestamp', async () => {
    const wrapper = createWrapper({
      stepStates: [
        {cursor: 0, expanded: true},
        {cursor: null, expanded: false},
        {cursor: null, expanded: false},
      ],
    });
    const logLines = [
      {index: 1, timestamp: 1765163618, message: 'Starting build'},
    ];
    wrapper.vm.appendLogs(0, logLines, 1765163618);

    // pre-condition - expect log timestamps are hidden
    expect(wrapper.find('.log-time-stamp').exists()).toBe(true);
    expect(wrapper.find('.log-time-stamp').classes()).toContain('tw-hidden');

    await wrapper.setProps({timeVisibleTimestamp: true});

    expect(wrapper.find('.log-time-stamp').exists()).toBe(true);
    expect(wrapper.find('.log-time-stamp').classes()).not.toContain('tw-hidden');
  });

  test('toggle visibility of duration seconds', async () => {
    const wrapper = createWrapper({
      stepStates: [
        {cursor: 0, expanded: true},
        {cursor: null, expanded: false},
        {cursor: null, expanded: false},
      ],
    });
    const logLines = [
      {index: 1, timestamp: 1765163618, message: 'Starting build'},
    ];
    wrapper.vm.appendLogs(0, logLines, 1765163618);

    // pre-condition - expect log time seconds are hidden
    expect(wrapper.find('.log-time-seconds').exists()).toBe(true);
    expect(wrapper.find('.log-time-seconds').classes()).toContain('tw-hidden');

    await wrapper.setProps({timeVisibleSeconds: true});

    expect(wrapper.find('.log-time-seconds').exists()).toBe(true);
    expect(wrapper.find('.log-time-seconds').classes()).not.toContain('tw-hidden');
  });

  test('emits toggle event on click when expandable', async () => {
    const wrapper = createWrapper();
    await wrapper.find('.job-step-summary').trigger('click');
    expect(wrapper.emitted('toggleStepLogs')).toBeTruthy();
    expect(wrapper.emitted('toggleStepLogs')).toHaveLength(1);
    expect(wrapper.emitted('toggleStepLogs')).toStrictEqual([[0]]); // step index that was toggled
  });

  test('scrollIntoView focuses on a line from the log', () => {
    const wrapper = createWrapper({
      stepStates: [
        {cursor: 0, expanded: true},
        {cursor: null, expanded: false},
        {cursor: null, expanded: false},
      ],
    });
    const logLines = [];
    for (let i = 0; i < 1000; i++) {
      logLines.push({index: 1 + i, timestamp: 1765163618 + i, message: `Line ${i}`});
    }
    wrapper.vm.appendLogs(0, logLines, 1765163618);

    const scrollIntoViewMock = vi.fn();
    const targetElement = wrapper.find('#jobstep-0-999 .line-num').element;
    targetElement.scrollIntoView = scrollIntoViewMock;

    wrapper.vm.scrollIntoView(0, '#jobstep-0-999');

    expect(scrollIntoViewMock).toHaveBeenCalled();
  });
});
