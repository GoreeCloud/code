// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

import {flushPromises, mount} from '@vue/test-utils';
import RepoActivityTopAuthors from './RepoActivityTopAuthors.vue';
import {expect, test, vi} from 'vitest';

test('calc image size and shift', async () => {
  vi.spyOn(RepoActivityTopAuthors.methods, 'init').mockResolvedValue({});

  const repoActivityTopAuthors = mount(RepoActivityTopAuthors, {
    props: {
      locale: {
        commitActivity: '',
      },
    },
  });
  await flushPromises();

  const square = repoActivityTopAuthors.vm.calcImageSizeAndShift({naturalWidth: 50, naturalHeight: 50});
  expect(square).toEqual([20, 20, 0, 0]);

  const portrait = repoActivityTopAuthors.vm.calcImageSizeAndShift({naturalWidth: 5, naturalHeight: 50});
  expect(portrait).toEqual([2, 20, 9, 0]);

  const landscape = repoActivityTopAuthors.vm.calcImageSizeAndShift({naturalWidth: 500, naturalHeight: 5});
  expect(landscape).toEqual([20, 0.2, 0, 9.9]);
});
