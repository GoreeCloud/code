// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// modules/markup/**
// web_src/js/features/repo-unicode-escape.js
// @watch end

import {expect} from '@playwright/test';
import {dynamic_id, test} from './utils_e2e.ts';

test.use({user: 'user2'});

test('Escape button in file preview', async ({page}) => {
  await page.goto('/user2/unicode-escaping/src/branch/main/a-file');

  const url = await page.getByRole('link', {name: 'Permalink'}).getAttribute('href');

  const response = await page.goto('/user2/repo1/issues/new');
  expect(response?.status()).toBe(200);

  // Create a new issue
  await page.getByPlaceholder('Title').fill(dynamic_id());
  await page.getByPlaceholder('Leave a comment').fill(`http://localhost:3003${url}#L1`);
  await page.getByRole('button', {name: 'Create issue'}).click();

  await expect(page).toHaveURL(/\/user2\/repo1\/issues\/\d+$/);

  await expect(page.locator('table.file-preview.unicode-escaped')).toHaveCount(0);
  await expect(async () => {
    await page.locator('button.toggle-escape-button').click();
    await expect(page.locator('table.file-preview.unicode-escaped')).toHaveCount(1);
  }).toPass();
});
