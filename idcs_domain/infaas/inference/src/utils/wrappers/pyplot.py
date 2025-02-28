import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import glob
import re

from matplotlib.pyplot import Axes, Figure
from typing import Optional, List, Any


class BenchmarkPlotUtils:

    def __init__(
        self,
        num_tests: int,
        data_path: str
    ) -> None:

        self._bar_width = 0.2
        self._num_tests = num_tests
        self._data_path = data_path

        csv_files = glob.glob(f"{self._data_path}/*.csv")
        print(f"{len(csv_files)} csvs were found")

        data_dfs = [pd.read_csv(file) for file in csv_files]
        data_dfs = [df[df['http_code'] == 200] for df in data_dfs]
        data_dfs = [
            df.apply(pd.to_numeric, errors='coerce')
            for df in data_dfs
        ]

        threads = [
            int(re.search(r'-(\d+)t', file).group(0)[1:-1])
            for file in csv_files
        ]

        zipped = zip(threads, data_dfs)
        sorted_zipped = sorted(zipped, key=lambda x: x[0])
        self._threads, self._data_dfs = zip(*sorted_zipped)

    def plot_percentiles_bar(
        self,
        data_header: str,
        title: str,
        ytitle: Optional[str] = "",
        xtitle: Optional[str] = "",
        colors: Optional[List[str]] = ["b", "g", "m", "orange"],
        percentiles: List[float] = [0.5, 0.75, 0.9, 0.99]
    ) -> None:

        percentiles_dfs = [
            df[data_header].quantile(percentiles)
            for df in self._data_dfs
        ]

        fig, ax = plt.subplots()

        for i, p in enumerate(percentiles):

            plots = [df[p] for df in percentiles_dfs]
            positions = np.arange(len(self._threads)) + (i * self._bar_width)

            ax.bar(
                positions,
                plots,
                color=colors[i],
                width=self._bar_width,
                label=f"p-{p}"
            )

        self._basic_bar_setup(
            yaxis_list=percentiles,
            xaxis_list=self._threads,
            title=title,
            xtitle=xtitle,
            ytitle=ytitle,
            fig=fig,
            ax=ax
        )

    def plot_completed_bar(
        self,
        title: str,
        ytitle: Optional[str] = "",
        xtitle: Optional[str] = ""
    ) -> None:

        # self._data_dfs is already filtered for response code 200 only
        completed_list = [df.shape[0] for df in self._data_dfs]
        sent_list = [t * self._num_tests for t in self._threads]
        completed_ratio = [
            round(c / s, 2)
            for c, s in zip(completed_list, sent_list)
        ]

        fig, ax = plt.subplots()
        ax.bar(
            self._threads,
            completed_ratio,
            color="b",
            width=self._bar_width * 3,
            label="c/s"
        )

        self._basic_bar_setup(
            xaxis_list=self._threads,
            title=title,
            xtitle=xtitle,
            ytitle=ytitle,
            fig=fig,
            ax=ax
        )

    def _basic_bar_setup(
        self,
        xaxis_list: List[Any],
        title: str,
        ytitle: str,
        xtitle: str,
        fig: Figure,
        ax: Axes,
        yaxis_list: Optional[List[Any]] = None
    ) -> None:

        if yaxis_list is not None:
            xticks = np.arange(len(xaxis_list)) +\
                     (len(yaxis_list) - 1) * self._bar_width / 2
        else:
            xticks = xaxis_list

        ax.set_xticks(xticks)
        ax.set_xticklabels(xaxis_list)
        ax.set_xlabel(xtitle)
        ax.set_ylabel(ytitle)

        plt.title(title, weight="bold")
        fig.tight_layout()
        fig.legend()

        plt.grid()
        plt.show()
