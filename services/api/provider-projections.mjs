export function projectRepository(data, owner, name) {
  if (!data || typeof data.full_name !== 'string' || data.full_name.toLowerCase() !== (owner + '/' + name).toLowerCase()) throw Error('repository identity mismatch');
  return {fullName: data.full_name, description: typeof data.description === 'string' ? data.description.slice(0, 1000) : '', private: data.private === true, archived: data.archived === true, defaultBranch: typeof data.default_branch === 'string' ? data.default_branch.slice(0,255) : null};
}

export function projectCollection(kind, data) {
  if (!Array.isArray(data)) throw Error('provider collection invalid');
  const items = data.slice(0, 20).map(item => {
    if (!item || typeof item !== 'object') throw Error('invalid provider item');
    if (kind === 'branches') {
      if (typeof item.name !== 'string' || !item.name) throw Error('invalid branch');
      return { name: item.name.slice(0,255), protected: item.protected === true,
        commitId: typeof item.commit?.id === 'string' ? item.commit.id.slice(0,64) : null };
    }
    if (kind === 'issues' || kind === 'pulls') {
      if (!Number.isSafeInteger(item.number) || item.number < 1 ||
          typeof item.title !== 'string') throw Error('invalid work item');
      return { number: item.number, title: item.title.slice(0,500),
        state: item.state === 'open' || item.state === 'closed' ? item.state : 'unknown',
        author: typeof item.user?.login === 'string' ? item.user.login.slice(0,100) : null,
        ...(kind === 'pulls' ? { draft: item.draft === true, merged: item.merged === true } : {}) };
    }
    throw Error('unsupported provider collection');
  });
  return { items, count: items.length };
}
