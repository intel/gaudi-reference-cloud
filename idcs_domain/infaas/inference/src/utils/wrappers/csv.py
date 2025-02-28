import csv
import re

from datetime import date
from typing import Optional, List, Dict


class CsvWriter:

    @staticmethod
    def write(
        data: Dict[str, List[Dict[str, str]]],
        headers: List[str],
        write_dir: str,
        file_name: str,
        add_date_to_file_name: Optional[bool] = True
    ) -> None:

        file_path = f"{write_dir}/{file_name}.csv"
        if add_date_to_file_name:
            today = date.today().strftime("%Y%m%d")
            file_path = re.sub(".csv", f"-{today}.csv", file_path)

        with open(file_path, "w") as csvfile:

            writer = csv.DictWriter(csvfile, fieldnames=headers)
            writer.writeheader()

            data_as_str_list = []
            for key in data.keys():
                res_dicts = data[key]
                for dict_row in res_dicts:
                    new_entry = [key]
                    dict_row_values_str = [
                        CsvWriter._get_csv_safe_str(v)
                        for v in dict_row.values()
                    ]
                    new_entry.extend(dict_row_values_str)
                    data_as_str_list.append(",".join(new_entry))

            data_as_csv_str = "\n".join(data_as_str_list)
            csvfile.write(data_as_csv_str)

    @staticmethod
    def _get_csv_safe_str(src: str) -> str:
        src = re.sub(",", ";", str(src))
        src = re.sub("\n", " ", src)
        return src
