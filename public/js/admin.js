var app = new Vue({
  el: '#app',
  data: {
    heading: "Admin Page",
    prefix: "",
    current: null,
    create: false,
    saving: false,
    m: {
      import_root: null,
      vcs_root: null,
      vcs: null,
      suffix: null,
    },
    repos: null
  },
  created: function () {
    this.updateRepos();
  },
  methods: {
    deleteRepo: function (evt) {
      target = evt.srcElement.attributes['data-repo'].value;
      if (target) {
        if (confirm("Delete " + target + "?")) {
          var me = this;
          var x = axios.create({});

          x.delete("_api/" + target)
            .then(function (res) {
              console.log("DELETED", target, res);
              me.updateRepos();
            })
            .catch(function (err) {
              console.log("ERROR", err.response.data);
            });
        }
      }
    },
    updateRepos: function () {
      var me = this;
      var x = axios.create({});

      me.current = null;
      x.get("_api/?prefix=" + me.prefix)
        .then(function (res) {
          me.repos = res.data;
        })
        .catch(function (err) {
          console.log("ERROR", err.response.data);
        });
    },
    createRepo: function () {
      var me = this;
      var x = axios.create({});

      me.saving = true;
      x.post("_api/", me.m)
        .then(function (res) {
          me.saving = false;
          me.create = false;
          me.m = {
            import_root: null,
            vcs_root: null,
            vcs: null,
            suffix: null
          };
          me.updateRepos();
        })
        .catch(function (err) {
          me.saving = false;
          console.log("ERROR", err.response.data);
        });
    },
    createValid: function () {
      return this.m.import_root && this.m.vcs_root && this.m.vcs;
    }
  }
});
