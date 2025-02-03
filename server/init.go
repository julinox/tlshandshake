package server

import (
	"tlesio/systema"
	iff "tlesio/tlssl/interfaces"
	mx "tlesio/tlssl/modulos"

	clog "github.com/julinox/consolelogrus"
	"github.com/sirupsen/logrus"
)

type zzl struct {
	lg   *logrus.Logger
	mods mx.TLSModulo
	ifs  *iff.Interfaces
	m0dz *mx.Modulos
}

func initTLS() (*zzl, error) {

	var ssl zzl
	var err error

	ssl.lg = getTLSLogger()
	ssl.m0dz = mx.NewModulos()
	ssl.m0dz.Cert, err = initModCerts(ssl.lg)
	if err != nil {
		ssl.lg.Error("initi certificates module: ", err)
		return nil, err
	}

	ssl.lg.Info("Module loaded: ", ssl.m0dz.Cert.Name())
	pp, err1 := ssl.m0dz.Cert.GetByCriteria(mx.CriterionCN("julINoX.cOm.ar"))
	if err1 != nil {
		ssl.lg.Error("error getting certificate: ", err1)
	} else {
		ssl.lg.Info("CertificateX: ", pp.Subject.CommonName)
	}

	return &ssl, nil
}

func initTLS2() (*zzl, error) {

	var ssl zzl
	var err error

	ssl.lg = getTLSLogger()
	ssl.mods, err = mx.InitModulos(ssl.lg, getTLSModules(ssl.lg))
	if err != nil {
		ssl.lg.Error("error initializing modules: ", err)
		return nil, err
	}

	ssl.ifs, err = iff.InitInterfaces(ssl.lg, ssl.mods)
	if err != nil {
		ssl.lg.Error("error initializing handshake interfaces: ", err)
		return nil, err
	}

	if ssl.mods == nil || ssl.ifs == nil {
		return nil, systema.ErrNilModulo
	}

	return &ssl, nil
}

func getTLSLogger() *logrus.Logger {

	lg := clog.InitNewLogger(&clog.CustomFormatter{
		Tag: "TLS", TagColor: "blue"})
	if lg == nil {
		return nil
	}

	lg.SetLevel(systema.GetLogLevel(_ENV_LOG_LEVEL_VAR_))
	return lg
}

func getTLSModules(lg *logrus.Logger) []mx.ModuloInfo {

	var basicModules []mx.ModuloInfo

	basicModules = append(basicModules, getModuleCertificateLoad(lg))
	basicModules = append(basicModules, getModuleCipherSuites())
	basicModules = append(basicModules, getModuleSignAlgo())
	return basicModules
}

func getModuleCertificateLoad(lg *logrus.Logger) mx.ModuloInfo {

	return mx.ModuloInfo{
		Id: 0xFFFE,
		Fn: mx.ModuloCertificates,
		Config: mx.CertificatesConfig{
			Lg: lg,
			Certs: []mx.CertificatesData_1{{
				PathCert: "./certs/server.crt",
				PathKey:  "./certs/server.key",
			}, {
				PathCert: "./certs/server2.crt",
				PathKey:  "./certs/server.key"},
			},
		},
	}
}

func getModuleCipherSuites() mx.ModuloInfo {

	return mx.ModuloInfo{
		Id: 0xFFFF,
		Fn: mx.ModuloCipherSuites,
	}
}

func getModuleSignAlgo() mx.ModuloInfo {

	return mx.ModuloInfo{
		Id: 0x000D,
		Fn: mx.ModuloSignAlgo,
	}
}

func initModCerts(lg *logrus.Logger) (mx.ModCerts, error) {

	if lg == nil {
		return nil, systema.ErrNilLogger
	}

	certs := []*mx.CertPaths{
		{PathCert: "./certs/server.crt", PathKey: "./certs/server.key"},
		{PathCert: "./certs/server2.crt", PathKey: "./certs/server.key"},
	}

	return mx.NewModCerts(lg, certs)
}
