// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// web_src/js/markup/external.js
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test('iframe renderer shrinks to shorter page', async ({page}, _workerInfo) => {
  const previewPath = '/user2/rendering-test/src/branch/master/short.iframehtml';

  const response = await page.goto(previewPath, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const preview = page.locator('iframe.external-render');
  await expect.poll(async () => {
    const boundingBox = await preview.boundingBox();
    return boundingBox.height;
  }).toBeLessThan(300);
});

test('iframe renderer expands to taller page', async ({page}, _workerInfo) => {
  const previewPath = '/user2/rendering-test/src/branch/master/tall.iframehtml';

  const response = await page.goto(previewPath, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const preview = page.locator('iframe.external-render');
  await expect.poll(async () => {
    const boundingBox = await preview.boundingBox();
    return boundingBox.height;
  }).toBeGreaterThan(300);
});

test('iframe renderer expands to taller page with absolutely-positioned body', async ({page}, _workerInfo) => {
  const previewPath = '/user2/rendering-test/src/branch/master/absolute.iframehtml';

  const response = await page.goto(previewPath, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const preview = page.locator('iframe.external-render');
  await expect.poll(async () => {
    const boundingBox = await preview.boundingBox();
    return boundingBox.height;
  }).toBeGreaterThan(300);
});

test('iframe renderer remains at default height if script breaks', async ({page}, _workerInfo) => {
  const previewPath = '/user2/rendering-test/src/branch/master/fail.iframehtml';

  const response = await page.goto(previewPath, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const preview = page.locator('iframe.external-render');
  await expect.poll(async () => {
    const boundingBox = await preview.boundingBox();
    return boundingBox.height;
  }).toBeCloseTo(300, 0.5);
});
