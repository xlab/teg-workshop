#include "gd.h"
#include "poly.h"
#include "serie.h"
#include "smatrix.h"
#include "tools.h"
#include "vars.h"

#include "bridge.h"

void* newGd(int g, int d) {
	return new gd;
}

void* newPoly() {
	return new poly;
}

void* newSerie(void* p, void* q, void* r) {
	serie* s = new serie;
	s->init(*(poly*)p, *(poly*)q, *(gd*)r);
	return s;
}

void freeGd(void* m) {
	delete (gd*)m;
}

void freePoly(void* p) {
	delete (poly*)p;
}

void freeSerie(void* s) {
	delete (serie*)s;
}

void appendPoly(void* p, int g, int d) {
	gd m(g, d);
	((poly*)p)->add(m);
}

void simplyPoly(void* p) {
	((poly*)p)->simpli();
}

unsigned int lenPoly(void* p) {
	return ((poly*)p)->getn();
}

void* getPoly(void* p, int n) {
	return &((poly*)p)->getpol(n);
}

int getG(void* m) {
	return ((gd*)m)->getg();
}

int getD(void* m) {
	return ((gd*)m)->getd();
}

void* getP(void* s) {
	return &((serie*)s)->getp();
}

void* getQ(void* s) {
	return &((serie*)s)->getq();
}

void* getR(void* s) {
	return &((serie*)s)->getr();
}

void* starPoly(void* p) {
	serie s(*(poly*)p);
	return new serie(star(s));
}

void* starSerie(void* s) {
	return new serie(star(*(serie*)s));
}

void* oplusPoly(void* p1, void* p2) {
	return new poly(oplus(*(poly*)p1, *(poly*)p2));
}

void* otimesPoly(void* p1, void* p2) {
	return new poly(otimes(*(poly*)p1, *(poly*)p2));
}

void* oplusSerie(void* s1, void* s2) {
	return new serie(oplus(*(serie*)s1, *(serie*)s2));
}

void* otimesSerie(void* s1, void* s2) {
	return new serie(otimes(*(serie*)s1, *(serie*)s2));
}
