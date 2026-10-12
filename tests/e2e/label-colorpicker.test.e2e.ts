// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// web_src/js/features/colorpicker.js
// templates/repo/issue/labels/label_new.tmpl
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test.use({user: 'user2'});

test('Color picker is visible above the new label dialog', async ({page}) => {
  const response = await page.goto('/user2/repo1/labels');
  expect(response?.status()).toBe(200);

  // Open the "New label" dialog
  await page.getByRole('button', {name: 'New label'}).click();
  const dialog = page.locator('dialog#new-label-modal');
  await expect(dialog).toBeVisible();

  // Click the color preview square to open the color picker popup
  const colorInput = dialog.locator('.js-color-picker-input input[name="color"]');
  await colorInput.click();

  // The hex-color-picker should be visible and not hidden behind the dialog
  const picker = dialog.locator('hex-color-picker');
  await expect(picker).toBeVisible();

  // Verify the picker tippy popup is a descendant of the dialog element,
  // not appended to document.body (which would render below the dialog top layer)
  const pickerInDialog = dialog.locator('.tippy-box hex-color-picker');
  await expect(pickerInDialog).toBeVisible();
});
