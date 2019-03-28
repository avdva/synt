// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

type mutexChecker struct {
	reports []CheckReport
}

func newMutexChecker() *mutexChecker {
	return &mutexChecker{}
}

func (sc *mutexChecker) DoPackage(info *CheckInfo) ([]CheckReport, error) {
	desc, err := makePkgDesc(info.Pkg, info.Fs)
	if err != nil {
		return nil, err
	}
	var results []CheckReport
	results = append(results, sc.checkGlobals(desc)...)
	return nil, nil
}

func (mc *mutexChecker) checkGlobals(desc *scopeDefs) []CheckReport {

	return nil
}
