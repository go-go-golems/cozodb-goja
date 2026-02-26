const cozo = require("cozodb");

module.exports = {
  apiVersion: "cozo.extractor/v1",
  kind: "extractor",
  id: "probe.create-no-replace",
  name: "probe create no replace",
  create() {
    return {
      run(input) {
        const o = input.engineOptions || {};
        const db = cozo.open({
          backend: o.backend || "cozo_cgo",
          engine: o.engine || "sqlite",
          path: o.path || "/tmp/cozo-probe3.db",
          options: o.options || {},
        });
        const pCreate = db.rel("users").create({ Keys: { id: "String" }, Values: { name: "String" } }, { Replace: false });
        const pPut = db.rel("users").put([{ id: "u1", name: "Ada" }], { Returning: true });
        const pExport = db.export(["users"]);
        const pClose = db.close();
        return { create: pCreate, put: pPut, export: pExport, close: pClose };
      },
    };
  },
};
