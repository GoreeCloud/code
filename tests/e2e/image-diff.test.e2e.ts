// @watch start
// templates/repo/diff/**
// web_src/css/features/imagediff.css
// web_src/css/modules/tab.css
// web_src/js/modules/tab.ts
// @watch end

import {expect} from '@playwright/test';
import {test, dynamic_id} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

test.use({user: 'user2'});

test('Repository image diff', async ({page}) => {
  // Generate a temporary SVG and edit it.
  let response = await page.goto('/user2/repo1/_new/master', {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  const filename = `${dynamic_id()}.svg`;

  await page.getByPlaceholder('Name your file…').fill(filename);
  await page.locator('.cm-content').click();
  await page.keyboard.type('<svg version="1.1" width="300" height="200" xmlns="http://www.w3.org/2000/svg"><circle cx="150" cy="100" r="80" fill="green" /></svg>\n');

  await page.locator('.quick-pull-choice input[value="direct"]').click();
  await page.getByRole('button', {name: 'Commit changes'}).click();

  response = await page.goto(`/user2/repo1/_edit/master/${filename}`, {waitUntil: 'domcontentloaded'});
  expect(response?.status()).toBe(200);

  await page.locator('.cm-content').click();
  await page.keyboard.press('Meta+KeyA');
  await page.keyboard.type('<svg version="1.1" width="300" height="200" xmlns="http://www.w3.org/2000/svg"><circle cx="150" cy="100" r="80" fill="red" /></svg>\n');

  await page.locator('.quick-pull-choice input[value="direct"]').click();
  await page.getByRole('button', {name: 'Commit changes'}).click();

  // Go to the commit page, where a image diff is shown.
  await page.locator('.commit-summary a.default-link').click();

  // Exhaustively test tabs works as expected
  await expect(page.locator('.item[data-tab="diff-side-by-side-1"]')).toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-swipe-1"]')).not.toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-overlay-1"]')).not.toContainClass('active');
  await expect(page.locator('.tab[data-tab="diff-side-by-side-1"]')).toBeVisible();
  await expect(page.locator('.tab[data-tab="diff-swipe-1"]')).toBeHidden();
  await expect(page.locator('.tab[data-tab="diff-overlay-1"]')).toBeHidden();
  await screenshot(page, page.locator('#diff-container'));

  await page.getByText('Swipe').click();
  await expect(page.locator('.item[data-tab="diff-side-by-side-1"]')).not.toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-swipe-1"]')).toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-overlay-1"]')).not.toContainClass('active');
  await expect(page.locator('.tab[data-tab="diff-side-by-side-1"]')).toBeHidden();
  await expect(page.locator('.tab[data-tab="diff-swipe-1"]')).toBeVisible();
  await expect(page.locator('.tab[data-tab="diff-overlay-1"]')).toBeHidden();
  await screenshot(page, page.locator('#diff-container'));

  await page.getByText('Overlay').click();
  await expect(page.locator('.item[data-tab="diff-side-by-side-1"]')).not.toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-swipe-1"]')).not.toContainClass('active');
  await expect(page.locator('.item[data-tab="diff-overlay-1"]')).toContainClass('active');
  await expect(page.locator('.tab[data-tab="diff-side-by-side-1"]')).toBeHidden();
  await expect(page.locator('.tab[data-tab="diff-swipe-1"]')).toBeHidden();
  await expect(page.locator('.tab[data-tab="diff-overlay-1"]')).toBeVisible();
  await screenshot(page, page.locator('#diff-container'));
});
