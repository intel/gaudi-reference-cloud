#!/bin/bash
# Copyright (c) 2024 Habana Labs, Ltd. an Intel Company.Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# LOG_DIR=logs/$(date +'%m-%Y/%m-%d-%Y/%m-%d-%Y_%H-%M')

LOG_DIR=hccl_demo_logs/
python3 screen.py --initialize --logs-dir $LOG_DIR;
python3 screen.py --screen --logs-dir $LOG_DIR;
