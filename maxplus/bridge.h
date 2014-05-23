#ifndef BRIDGE_H
#define BRIDGE_H

#ifdef __cplusplus
extern "C" {
#endif

typedef void gd_;
typedef void poly_;
typedef void serie_;

gd_ *newGd(int g, int d);
poly_ *newPoly();
serie_ *newSerie(poly_ *p, poly_ *q, gd_ *r);

void freeGd(gd_ *p);
void freePoly(poly_ *p);
void freeSerie(serie_ *s);

void appendPoly(poly_ *p, int g, int d);
void simplyPoly(poly_ *p);
unsigned int lenPoly(poly_ *p);
gd_ *getPoly(poly_ *p, int n);

int getG(gd_ *m);
int getD(gd_ *m);
poly_ *getP(poly_ *s);
poly_ *getQ(poly_ *s);
gd_ *getR(gd_ *s);

poly_ *oplusPoly(poly_ *p1, poly_ *p2);
poly_ *otimesPoly(poly_ *p1, poly_ *p2);
serie_ *oplusSerie(serie_ *s1, serie_ *s2);
serie_ *otimesSerie(serie_ *s1, serie_ *s2);
serie_ *starPoly(poly_ *p);
serie_ *starSerie(serie_ *s);

#ifdef __cplusplus
}
#endif

#endif