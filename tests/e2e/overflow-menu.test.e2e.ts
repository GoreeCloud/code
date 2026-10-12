// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

// @watch start
// templates/user/profile.tmpl
// templates/user/overview/header.tmpl
// web_src/css/modules/tippy.css
// web_src/js/modules/tippy.js
// @watch end

import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';

test(`Visual properties`, async ({page, isMobile}) => {
  test.skip(!isMobile, 'Overflow menu button only appears on mobile');

  const noBg = 'rgba(0, 0, 0, 0)';
  const activeBg = 'rgb(226, 226, 229)';
  const menuItemSelector = `.tippy-box .tippy-content .tippy-target > a.item`;
  const activeItemSelector = `${menuItemSelector}.active`;
  const inactiveItemSelector = `${menuItemSelector}:not(.active)`;

  await page.goto(`/user2/repo1`);
  const overflowMenuButton = page.locator(`.overflow-menu-button`);

  await overflowMenuButton.click();
  const menuItems = page.locator(`${menuItemSelector}`);
  const itemCount = await menuItems.count();
  for (let i = 0; i < itemCount; i++) {
    await menuItems.nth(i).click();
    await page.waitForLoadState('domcontentloaded');

    await overflowMenuButton.click();
    const activeItem = page.locator(`${activeItemSelector}`);
    expect(await activeItem.evaluate((el) => getComputedStyle(el).backgroundColor)).toBe(activeBg);

    const inactiveItems = page.locator(inactiveItemSelector);
    const inactiveCount = await inactiveItems.count();
    for (let j = 0; j < itemCount - inactiveCount; j++) {
      const nonActiveItem = page.locator(`${inactiveItemSelector}`).nth(j);
      expect(await nonActiveItem.evaluate((el) => getComputedStyle(el).backgroundColor)).toBe(noBg);
    }
  }
});
