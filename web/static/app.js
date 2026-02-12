function app() {
  return {
    // State
    view: 'stats',
    viewId: null,
    loading: false,
    stats: null,
    searchQuery: '',
    searchResults: [],

    // Persons
    personsList: [],
    personsPage: 1,
    personsPerPage: 100,
    personsTotal: 0,
    surnameFilter: '',
    personDetail: null,
    activePanel: null, // 'ancestors', 'descendants', 'treetops', 'summary'
    ancestorsList: [],
    descendantsList: [],
    treetopsList: [],
    summaryData: null,

    // Families
    familiesList: [],
    familiesPage: 1,
    familiesPerPage: 100,
    familiesTotal: 0,
    familyDetail: null,

    // Places
    placesList: [],
    placeDetail: null,
    placePersonsList: [],

    // Events
    eventsList: [],
    eventDetail: null,
    eventPersonsList: [],

    // Sources
    sourcesList: [],
    sourceDetail: null,
    sourcePersonsList: [],

    async init() {
      // Load stats on init
      this.stats = await this.api('/api/stats');
      // Handle hash routing
      window.addEventListener('hashchange', () => this.handleHash());
      this.handleHash();
    },

    handleHash() {
      const hash = window.location.hash.slice(1); // remove #
      if (!hash || hash === '/') {
        this.navigate('stats');
        return;
      }
      const parts = hash.split('/').filter(Boolean);
      if (parts.length >= 2) {
        this.navigate(parts[0], parseInt(parts[1], 10));
      } else if (parts.length === 1) {
        this.navigate(parts[0]);
      }
    },

    async navigate(view, id) {
      this.view = view;
      this.viewId = id || null;
      this.activePanel = null;
      this.ancestorsList = [];
      this.descendantsList = [];
      this.treetopsList = [];
      this.summaryData = null;

      // Update hash without triggering hashchange
      const newHash = id ? `#${view}/${id}` : `#${view}`;
      if (window.location.hash !== newHash) {
        history.pushState(null, '', newHash);
      }

      switch (view) {
        case 'stats': await this.loadStats(); break;
        case 'persons': await this.loadPersons(); break;
        case 'person': await this.loadPersonDetail(id); break;
        case 'families': await this.loadFamilies(); break;
        case 'family': await this.loadFamilyDetail(id); break;
        case 'places': await this.loadPlaces(); break;
        case 'place': await this.loadPlaceDetail(id); break;
        case 'events': await this.loadEvents(); break;
        case 'event': await this.loadEventDetail(id); break;
        case 'sources': await this.loadSources(); break;
        case 'source': await this.loadSourceDetail(id); break;
      }
    },

    async api(url) {
      try {
        const resp = await fetch(url);
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        return await resp.json();
      } catch (e) {
        console.error('API error:', e);
        return null;
      }
    },

    formatPersonName(p) {
      if (!p) return '(unknown)';
      const g = p.given_name || '';
      const s = p.surname || '';
      if (!g && !s) return '(unknown)';
      return [g, s].filter(Boolean).join(' ');
    },

    sexLabel(sex) {
      switch (sex) {
        case 1: return 'M';
        case 2: return 'F';
        default: return 'U';
      }
    },

    async liveSearch() {
      if (this.searchQuery.length < 2) {
        this.searchResults = [];
        return;
      }
      const data = await this.api(`/api/search?q=${encodeURIComponent(this.searchQuery)}`);
      this.searchResults = data || [];
    },

    async loadStats() {
      this.loading = true;
      this.stats = await this.api('/api/stats');
      this.loading = false;
    },

    async loadPersons() {
      this.loading = true;
      let url = `/api/persons?page=${this.personsPage}&per_page=${this.personsPerPage}`;
      if (this.surnameFilter) url += `&surname=${encodeURIComponent(this.surnameFilter)}`;
      const data = await this.api(url);
      if (data) {
        this.personsList = data.items || [];
        this.personsTotal = data.total || 0;
      }
      this.loading = false;
    },

    async loadPersonDetail(id) {
      this.loading = true;
      this.personDetail = await this.api(`/api/persons/${id}`);
      this.loading = false;
    },

    clearPanels() {
      this.ancestorsList = [];
      this.descendantsList = [];
      this.treetopsList = [];
      this.summaryData = null;
    },

    async togglePanel(panel) {
      if (this.activePanel === panel) {
        this.activePanel = null;
        this.clearPanels();
        return;
      }
      this.clearPanels();
      this.activePanel = panel;
      switch (panel) {
        case 'ancestors': await this.loadAncestors(); break;
        case 'descendants': await this.loadDescendants(); break;
        case 'treetops': await this.loadTreetops(); break;
        case 'summary': await this.loadSummary(); break;
      }
    },

    async loadAncestors() {
      if (!this.personDetail) return;
      this.ancestorsList = await this.api(`/api/persons/${this.personDetail.id}/ancestors?generations=10`) || [];
    },

    async loadDescendants() {
      if (!this.personDetail) return;
      this.descendantsList = await this.api(`/api/persons/${this.personDetail.id}/descendants?generations=10`) || [];
    },

    async loadTreetops() {
      if (!this.personDetail) return;
      this.treetopsList = await this.api(`/api/persons/${this.personDetail.id}/treetops`) || [];
    },

    async loadSummary() {
      if (!this.personDetail) return;
      this.summaryData = await this.api(`/api/persons/${this.personDetail.id}/summary`);
    },

    async loadFamilies() {
      this.loading = true;
      const data = await this.api(`/api/families?page=${this.familiesPage}&per_page=${this.familiesPerPage}`);
      if (data) {
        this.familiesList = data.items || [];
        this.familiesTotal = data.total || 0;
      }
      this.loading = false;
    },

    async loadFamilyDetail(id) {
      this.loading = true;
      this.familyDetail = await this.api(`/api/families/${id}`);
      this.loading = false;
    },

    async loadPlaces() {
      this.loading = true;
      this.placesList = await this.api('/api/places') || [];
      this.loading = false;
    },

    async loadPlaceDetail(id) {
      this.loading = true;
      this.placeDetail = await this.api(`/api/places/${id}`);
      this.placePersonsList = await this.api(`/api/places/${id}/persons`) || [];
      this.loading = false;
    },

    async loadEvents() {
      this.loading = true;
      this.eventsList = await this.api('/api/events') || [];
      this.loading = false;
    },

    async loadEventDetail(id) {
      this.loading = true;
      this.eventDetail = await this.api(`/api/events/${id}`);
      this.eventPersonsList = await this.api(`/api/events/${id}/persons`) || [];
      this.loading = false;
    },

    async loadSources() {
      this.loading = true;
      this.sourcesList = await this.api('/api/sources') || [];
      this.loading = false;
    },

    async loadSourceDetail(id) {
      this.loading = true;
      this.sourceDetail = await this.api(`/api/sources/${id}`);
      this.sourcePersonsList = await this.api(`/api/sources/${id}/persons`) || [];
      this.loading = false;
    },
  };
}
