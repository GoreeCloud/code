export function projectRepository(data, owner, name) {
  if (!data || typeof data.full_name !== 'string' || data.full_name.toLowerCase() !== (owner + '/' + name).toLowerCase()) throw Error('repository identity mismatch');
  return {fullName: data.full_name, description: typeof data.description === 'string' ? data.description.slice(0, 1000) : '', private: data.private === true, archived: data.archived === true, defaultBranch: typeof data.default_branch === 'string' ? data.default_branch.slice(0,255) : null};
}
