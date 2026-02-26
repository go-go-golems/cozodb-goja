const cozo = require("cozodb");

module.exports = {
  apiVersion: "cozo.extractor/v1",
  kind: "extractor",
  id: "cozo.probe.import",
  name: "Probe - import relation rows",
  create() {
    return {
      run(input) {
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

        const pImport = db.import({
          users: {
            Headers: ["id", "name"],
            Rows: [["u1", "Ada"], ["u2", "Bob"]],
          },
        });
        const pExport = db.export(["users"]);
        const pClose = db.close();

        return { create: pCreate, import: pImport, export: pExport, close: pClose };
      },
    };
  },
};
