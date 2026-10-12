// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/repo/diff/**
// web_src/css/review.css
// web_src/js/features/repo-diff.js
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test.use({user: 'user2'});

test('Expand Large diff', async ({page}) => {
  let response = await page.goto('/user2/huge-diff-test/src/branch/main-2', {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const commitUrl = await page.locator('.commit-summary a').getAttribute('href');

  response = await page.goto(commitUrl, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const loadDiff = page.locator('.diff-load-button');
  const expandDiff = page.locator('.code-expander-button');
  const diffLines = page.locator('.file-body tbody > tr');

  await expect(loadDiff).toBeVisible();

  loadDiff.click();
  await page.waitForLoadState('load');

  await expect(expandDiff).toHaveCount(167);
  await expect(diffLines).toHaveCount(1495);

  await expandDiff.first().click();
  await page.waitForLoadState('load');

  await expect(expandDiff).toHaveCount(166);
  await expect(diffLines).toHaveCount(1502);
});
