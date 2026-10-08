// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/repo/commit_header.tmpl
// @watch end

import {expect} from '@playwright/test';
import {dynamic_id, test} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

test.use({user: 'user2'});

test('Create branch from commit', async ({page}) => {
  let response = await page.goto('/user2/repo1/commit/65f1bf27bc3bf70f64657658635e66094edbcb4d');
  expect(response?.status()).toBe(200);

  // Open create branch modal.
  await page.locator('.commit-header-buttons .dropdown.button').click();
  await page.getByRole('option', {name: 'Create branch'}).click();
  await expect(page.locator('#create-branch-modal')).toBeVisible();
  await screenshot(page, page.locator('#create-branch-modal'));

  // Check that it can be cancelled.
  await page.getByRole('button', {name: 'Cancel'}).click();
  await expect(page.locator('#create-branch-modal')).toBeHidden();

  // Open it again and make a branch.
  await page.locator('.commit-header-buttons .dropdown.button').click();
  await page.getByRole('option', {name: 'Create branch'}).click();
  await expect(page.locator('#create-branch-modal')).toBeVisible();

  const branchName = dynamic_id();
  await page.getByRole('textbox').fill(branchName);
  await page.getByRole('button', {name: 'Create branch'}).click();

  // Verify branch exists.
  response = await page.goto(`/user2/repo1/src/branch/${branchName}`);
  expect(response?.status()).toBe(200);
});

test('Create tag from commit', async ({page}) => {
  let response = await page.goto('/user2/repo1/commit/65f1bf27bc3bf70f64657658635e66094edbcb4d');
  expect(response?.status()).toBe(200);

  // Open create tag modal.
  await page.locator('.commit-header-buttons .dropdown.button').click();
  await page.getByRole('option', {name: 'Create tag'}).click();
  await expect(page.locator('#create-tag-modal')).toBeVisible();
  await screenshot(page, page.locator('#create-tag-modal'));

  // Check that it can be cancelled.
  await page.getByRole('button', {name: 'Cancel'}).click();
  await expect(page.locator('#create-tag-modal')).toBeHidden();

  // Open it again and make a branch.
  await page.locator('.commit-header-buttons .dropdown.button').click();
  await page.getByRole('option', {name: 'Create tag'}).click();
  await expect(page.locator('#create-tag-modal')).toBeVisible();

  const tagName = dynamic_id();
  await page.getByRole('textbox').fill(tagName);
  await page.getByRole('button', {name: 'Create tag'}).click();

  // Verify tag exists.
  response = await page.goto(`/user2/repo1/releases/tag/${tagName}`);
  expect(response?.status()).toBe(200);
});
