from abc import ABCMeta, abstractmethod

from cs.utils.result import AnalyticResult


class Analytic(metaclass=ABCMeta):
    @abstractmethod
    def compute(self): ...

    @abstractmethod
    def plot(self, data): ...

    @abstractmethod
    def display(self) -> AnalyticResult: ...
