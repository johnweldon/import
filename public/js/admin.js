var app = new Vue({
  el: '#app',
  data: {
    heading: "Admin Page",
    host: "jw4.us",
    current: null,
    create: false,
    saving: false,
    m: {
      import_root: null,
      vcs_root: null,
      vcs: null,
      suffix: null,
    },
    repos: []
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
          me.saving = true;
          var x = axios.create({
            headers: {
              'X-Host-Override': me.host
            }
          });
          x.delete("_api/" + target)
            .then(function (res) {
              console.log("DELETED", target, res);
              me.updateRepos();
            });
        }
      }
    },
    updateRepos: function () {
      var me = this;
      var x = axios.create({
        headers: {
          'X-Host-Override': me.host
        }
      });

      x.get("_api/")
        .then(function (res) {
          me.repos = res.data;
        })
        .catch(function (err) {
          console.log("ERROR", err.response.data);
        });
    },
    createRepo: function () {
      var me = this;
      me.saving = true;
      var x = axios.create({
        headers: {
          'X-Host-Override': me.host
        }
      });
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
    },
    createValid: function () {
      return this.m.import_root && this.m.vcs_root && this.m.vcs;
    }
  }
});
