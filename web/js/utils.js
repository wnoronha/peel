// Alpine component for peel. Loaded as a classic script (end of body) so it
// registers on alpine:init before the deferred Alpine bundle initializes.
const UI_TIMINGS = { resizeMs: 150, reloadMs: 1000, flashMs: 1400, toastMs: 3200, slideLockMs: 600 };

// utils.js — peel mixin (classic script, no modules).
function peelUtils() {
  return {
    typeIcon(t) {
      return t === 'c4' ? '[c4]' : (t === 'deck' ? '[dk]' : '[md]');
    },

    fuzzyScore(text, query) {
      const t = String(text || '').toLowerCase();
      const q = String(query || '').toLowerCase();
      if (!q) return 1;
      let ti = 0, score = 0, run = 0, first = -1;
      for (let qi = 0; qi < q.length; qi++) {
        const ch = q[qi];
        let found = -1;
        for (let j = ti; j < t.length; j++) {
          if (t[j] === ch) { found = j; break; }
        }
        if (found < 0) return 0;
        if (first < 0) first = found;
        run = found === ti ? run + 1 : 1;
        score += 8 + run * 6;
        if (found === 0 || /[\/\-_.]/.test(t[found - 1])) score += 18;
        ti = found + 1;
      }
      score += Math.max(0, 40 - first);
      score += Math.max(0, 30 - (t.length - q.length));
      return score;
    },
    fileParts(path) {
      const slash = path.lastIndexOf('/');
      if (slash < 0) return { name: path, dir: '' };
      return { name: path.slice(slash + 1), dir: path.slice(0, slash) };
    },
    timeAgo(ts) {
      const t = Date.parse(ts);
      if (isNaN(t)) return '';
      const s = Math.max(1, Math.floor((Date.now() - t) / 1000));
      if (s < 60) return s + 's ago';
      const m = Math.floor(s / 60);
      if (m < 60) return m + 'm ago';
      const h = Math.floor(m / 60);
      if (h < 24) return h + 'h ago';
      return Math.floor(h / 24) + 'd ago';
    },

  };
}

function wrapTables(root) {
  root.querySelectorAll('table').forEach(table => {
    if (table.closest('.table-scroll')) return;
    const wrap = document.createElement('div');
    wrap.className = 'table-scroll';
    table.replaceWith(wrap);
    wrap.appendChild(table);
  });
}

function mountDiagrams(root) {
  if (!window.pele) return;
  root.querySelectorAll('pre.diagram').forEach(pre => {
    const code = pre.querySelector('code');
    const src = (code ? code.textContent : pre.textContent).trim();
    if (/<script|javascript:/i.test(src)) {
      pre.textContent = 'diagram blocked: unsafe source';
      return;
    }
    const div = document.createElement('div');
    div.className = 'pele-diagram';
    pre.replaceWith(div);
    try { window.pele.mount(div, src); }
    catch (e) { div.textContent = 'diagram error: ' + e.message; }
  });
}

