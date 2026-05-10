export const frontendUser = {
  id: 1,
  username: 'admin',
  nickname: '超级管理员',
  is_moderator: true,
};

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
