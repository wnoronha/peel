// sse.js — peel mixin (classic script, no modules).
function peelSSE() {
  return {
    applyAgent(st) {
      if (!st || typeof st !== 'object') return;
      this.agent = {
        state: st.state === 'working' ? 'working' : 'idle',
        ref: st.ref || '',
        note: st.note || '',
        since: st.since || '',
      };
    },
    async loadAgent() {
      try {
        const r = await fetch('/api/agent');
        if (!r.ok) return;
        this.applyAgent(await r.json());
      } catch (_) { /* ignore */ }
    },

    toast(message, opts = {}) {
      const id = ++this._toastSeq;
      const kind = opts.kind || 'info'; // info | ok | warn
      const ttl = opts.ttl != null ? opts.ttl : 4500;
      const item = { id, message, kind, action: opts.action || null };
      this.toasts = [...this.toasts, item].slice(-4);
      if (ttl > 0) {
        setTimeout(() => this.dismissToast(id), ttl);
      }
    },

    dismissToast(id) {
      this.toasts = this.toasts.filter(t => t.id !== id);
    },

    onToastClick(t) {
      if (t.action === 'feedback') {
        this.rightOpen = true;
        localStorage.setItem('peel-right', '1');
        if (t.kind === 'ok') this.setFbStatus('acked');
        else this.setFbStatus('open');
        this.reloadFeedback();
      }
      this.dismissToast(t.id);
    },

    toastFromFeedback(ev) {
      if (!ev || typeof ev !== 'object') return;
      const t = ev.type || 'note';
      if (t === 'ack') {
        const detail = (ev.note || '').trim();
        const msg = detail
          ? ('Agent finished: ' + (detail.length > 72 ? detail.slice(0, 69) + '…' : detail))
          : 'Agent finished a note';
        this.toast(msg, { kind: 'ok', action: 'feedback' });
        return;
      }
      if (t === 'tombstone') {
        this.toast('Note dismissed', { kind: 'warn', action: 'feedback' });
        return;
      }
      if (t === 'note') {
        // Skip toasting our own saves (same tab still gets the event).
        if (ev.id && this._ownNoteIds && this._ownNoteIds[ev.id]) {
          delete this._ownNoteIds[ev.id];
          return;
        }
        const file = ev.file ? String(ev.file) : 'a file';
        const preview = (ev.note || ev.quote || '').trim();
        const bit = preview ? (preview.length > 48 ? preview.slice(0, 45) + '…' : preview) : file;
        this.toast('New note · ' + bit, { kind: 'info', action: 'feedback' });
      }
    },

    connectSSE() {
      const es = new EventSource('/events');
      es.onopen = () => { this.sse = 'live'; };
      es.onerror = () => { this.sse = 'retrying'; };
      es.onmessage = e => {
        try {
          const m = JSON.parse(e.data);
          if (m.type === 'reload') {
            const now = Date.now();
            if (!this._lastReload || now - this._lastReload > UI_TIMINGS.reloadMs) {
              this._lastReload = now;
              this.loadFiles();
            }
            if (m.path === this.current) {
              const hasDraft = this.composer && this.composer.open;
              if (!hasDraft && (!this._lastOpen || now - this._lastOpen > UI_TIMINGS.reloadMs)) {
                this._lastOpen = now;
                this.openFile(this.current);
              }
              this.toast('Doc updated · ' + m.path, { kind: 'info', ttl: UI_TIMINGS.toastMs });
            }
          } else if (m.type === 'feedback') {
            this.toastFromFeedback(m.event);
            this.reloadFeedback();
            this.refreshTotal();
          } else if (m.type === 'agent') {
            this.applyAgent(m);
          }
        } catch (_) { /* ignore */ }
      };
    },
  };
}
