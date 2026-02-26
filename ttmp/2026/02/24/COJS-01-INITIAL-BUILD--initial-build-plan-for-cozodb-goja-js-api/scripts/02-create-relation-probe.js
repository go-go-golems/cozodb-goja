const { defineExtractorPlugin, wrapExtractorRun } = require("geppetto/plugins");
const cozo = require("cozodb");

module.exports = defineExtractorPlugin({
  id: "cozo.probe.create-relation",
  name: "Probe - create relation",
  create() {
    return {
      run: wrapExtractorRun((input) => {
        const o = input.engineOptions || {};
        const db = cozo.open({
          backend: o.backend || "cozo_cgo",
          engine: o.engine || "sqlite",
          path: o.path || "/tmp/cozo-probe.db",
          options: o.options || {},
        });

        const users = db.rel("users");
        const pCreate = users.create(
          {
            Keys: { id: "String" },
            Values: { name: "String" },
          },
          { Replace: true }
        );
        const pPut = users.put([{ id: "u1", name: "Ada" }], { Returning: true });
        const pExport = db.export(["users"]);
        const pClose = db.close();

        return { create: pCreate, put: pPut, export: pExport, close: pClose };
      }),
    };
  },
});
