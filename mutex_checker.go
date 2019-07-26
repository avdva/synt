// Copyright 2017 Aleksandr Demakin. All rights reserved.

package synt

type mutexChecker struct {
	reports []CheckReport
}

func newMutexChecker() *mutexChecker {
	return &mutexChecker{}
}

func (mc *mutexChecker) DoPackage(info *CheckInfo) ([]CheckReport, error) {
	_, err := makePkgDesc(info.Pkg, info.Fs)
	if err != nil {
		return nil, err
	}
	var results []CheckReport
	results = append(results, mc.checkGlobals(nil)...)
	return nil, nil
}

func (mc *mutexChecker) checkGlobals(defs *defs) []CheckReport {

	return nil
}
