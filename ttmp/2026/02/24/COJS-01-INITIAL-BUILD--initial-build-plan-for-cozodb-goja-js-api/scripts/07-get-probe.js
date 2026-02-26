const cozo = require("cozodb");

module.exports = {
  apiVersion: "cozo.extractor/v1",
  kind: "extractor",
  id: "cozo.probe.get",
  name: "Probe - relation get",
  create() {
    return {
      run(input) {
        const o = input.engineOptions || {};
        const db = cozo.open({
          backend: o.backend || "cozo_cgo",
          engine: o.engine || "sqlite",
          path: o.path || "/tmp/cozo-js-examples.db",
          options: o.options || {},
        });

        const row = db.rel("users").get({ id: "u1" });
        const close = db.close();

        return { row, close };
      },
    };
  },
};
