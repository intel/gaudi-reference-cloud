#!/usr/bin/env python3
# authored by: michael vincerra | michael.vincerra@intel.com

import json
from datetime import datetime
from jinja2 import Template
from sphinx.application import Sphinx
from sphinx.util.docutils import SphinxDirective
from docutils.nodes import raw


class json2table(SphinxDirective):
    def run(self):
        ''' Load 'table_data_file', as defined in conf.py
            Populate contents via 'data' var to capture JSON file contents
            Automate creation of 'Software' table via custom directive in jupyter_learning
        '''
        table_data_file = self.env.config.table_data_file
        if table_data_file:
            with open(table_data_file, 'r') as f:
                data = json.load(f)
            template = Template(open('_templates/table_template.html').read())
            rendered_table = template.render(data=data, current_date=datetime.now().strftime('%x %H:%M'))
            return [raw('', rendered_table, format='html')]
        else:
            return []

def setup(app: Sphinx):
    app.add_config_value('table_data_file','', 'html')
    app.add_directive('render-json-table', json2table)