export function projectRepository(data, owner, name) {
  if (!data || typeof data.full_name !== 'string' || data.full_name.toLowerCase() !== (owner + '/' + name).toLowerCase()) throw Error('repository identity mismatch');
  return {fullName: data.full_name, description: typeof data.description === 'string' ? data.description.slice(0, 1000) : '', private: data.private === true, archived: data.archived === true, defaultBranch: typeof data.default_branch === 'string' ? data.default_branch.slice(0,255) : null};
}

export function projectCollection(kind, data) {
  if (!Array.isArray(data)) throw Error('provider collection invalid');
  return { items: data.slice(0,20).map(item => ({name: typeof item?.name === 'string' ? item.name.slice(0,255) : ''})) };
}
