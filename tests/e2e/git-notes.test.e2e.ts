// @ts-check
import {expect} from '@playwright/test';
import {test} from './utils_e2e.ts';
import {screenshot} from './shared/screenshots.ts';

test.use({user: 'user2'});

test('Change git note', async ({page}) => {
  const text = 'This is a new note <script>alert("xss")</script>.\nSee https://frogejo.org.';

  let response = await page.goto('/user2/repo1/commit/65f1bf27bc3bf70f64657658635e66094edbcb4d');
  expect(response?.status()).toBe(200);

  // An add button should not be present, because the commit already has a commit note
  await expect(page.locator('#commit-notes-add-button')).toHaveCount(0);

  let renderedarea = page.locator('#commit-notes-display-area pre.commit-body');
  await expect(renderedarea).toBeVisible();
  let textarea = page.locator('textarea[name="notes"]');
  await expect(textarea).toBeHidden();

  await page.locator('#commit-notes-edit-button').click();

  await expect(renderedarea).toBeHidden();
  await expect(textarea).toBeVisible();
  await textarea.fill(text);
  await screenshot(page, page.locator('.ui.container.fluid.padded'));

  await page.locator('#commit-notes-save-button').click();

  await expect(renderedarea).toBeVisible();
  await expect(textarea).toBeHidden();
  await expect(renderedarea).toHaveText(text);
  await expect(renderedarea.locator('a')).toHaveAttribute('href', 'https://frogejo.org');
  await screenshot(page, page.locator('.ui.container.fluid.padded'));

  // Check edited note
  response = await page.goto('/user2/repo1/commit/65f1bf27bc3bf70f64657658635e66094edbcb4d');
  expect(response?.status()).toBe(200);

  renderedarea = page.locator('#commit-notes-display-area pre.commit-body');
  await expect(renderedarea).toHaveText(text);
  await expect(renderedarea.locator('a')).toHaveAttribute('href', 'https://frogejo.org');
  textarea = page.locator('textarea[name="notes"]');
  await expect(textarea).toHaveText(text);
  await expect(textarea.locator('a')).toHaveCount(0);
  await screenshot(page, page.locator('.ui.container.fluid.padded'));

  // Cancel note editing
  await page.locator('#commit-notes-edit-button').click();
  await textarea.fill('Edited note');
  await page.locator('#commit-notes-cancel-button').click();
  await expect(renderedarea).toBeVisible();
  await expect(renderedarea).toHaveText(text);
  await expect(textarea).toBeHidden();
  await expect(textarea).toHaveText(text);
});
