import flask

app = flask.Flask(__name__)
app.config["DEBUG"] = True

def readData(path):
  f = open(path, "r")
  data = f.read()
  f.close()
  return data

@app.route('/kpimon/config', methods=['GET'])
def kpimon_config():
  return readData("kpimon/config.json")

@app.route('/kpimon/schema', methods=['GET'])
def kpimon_schema():
  return readData("kpimon/schema.json")

app.run(host='0.0.0.0', port=30030)
