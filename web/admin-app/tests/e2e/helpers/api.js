export async function ensurePluginEnabled(request, code) {
  await request.post(`/api/v1/admin/plugins/${code}/enable`, {
    headers: { Authorization: 'Bearer devhub-admin-1' },
  });
}
