export const frontendUser = {
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  is_moderator: false,
};

export const phpModeratorUser = {
  id: 2,
  username: 'operator',
  nickname: '运营管理员',
  is_moderator: true,
};

export const goModeratorUser = {
  id: 3,
  username: 'auditor',
  nickname: '内容审核员',
  is_moderator: true,
};

export async function clearFrontendAuth(page) {
  await page.goto('/');
  await page.evaluate(() => {
    localStorage.removeItem('devhub_user_token');
    localStorage.removeItem('devhub_user_refresh_token');
    localStorage.removeItem('devhub_access_token');
    localStorage.removeItem('devhub_user');
    window.dispatchEvent(new CustomEvent('devhub:session-change', { detail: { user: null, token: '' } }));
  });
}

export async function seedUserSession(page, user = frontendUser) {
  await page.goto('/');
  await page.evaluate((currentUser) => {
    localStorage.setItem('devhub_user_token', `devhub-user-${currentUser.id || 1}`);
    localStorage.setItem('devhub_user_refresh_token', `devhub-user-${currentUser.id || 1}-refresh`);
    localStorage.setItem('devhub_user', JSON.stringify(currentUser));
    window.dispatchEvent(new CustomEvent('devhub:session-change', {
      detail: { user: currentUser, token: `devhub-user-${currentUser.id || 1}` },
    }));
  }, user);
}

export async function loginAsUser(page) {
  await seedUserSession(page, frontendUser);
}

export async function loginAsModerator(page) {
  await seedUserSession(page, phpModeratorUser);
}

export async function loginAsAdmin(page) {
  await seedUserSession(page, frontendUser);
}
