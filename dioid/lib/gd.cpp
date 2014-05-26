#include "gd.h"

/*
 * Classe gd definissant un monome
 */

void gd::affecte(long g1, long d1)
{
    g = g1;
    d = d1;
}

gd::gd(void) // constructuer 0: (g,d)=(+00,-00)
{
    g = infty;
    d = -infty;
}

gd::gd(const gd& gd1) // constructeur 1: par recopie
{
    affecte(gd1.g, gd1.d);
}

gd::gd(long g1, long d1) // constructeur 2: initialise avec 2 entiers
{
    affecte(g1, d1);
}

gd& gd::operator=(const gd& gd1) // initialisation: par recopie
{
    if (&gd1 != this) {
        affecte(gd1.g, gd1.d); // ne rien faire si c'est le meme objet
    }
    return *this;
} // retourne l'objet courant pour permettre

// une affectation successive
int gd::operator!=(const gd& gd1) // comparaison d'un monome avec le monome courant
{
    if (gd1.g == g) {
        if (gd1.d == d)
        { return 0; }
    }
    return 1;
}

int gd::operator==(const gd& gd1) // comparaison d'un monome avec le monome courant
{
    if (gd1.g == g) {
        if (gd1.d == d)
        { return 1; }
    }
    return 0;
}

int gd::operator>=(const gd& gd1) // comparaison d'un monome avec le monome courant
{
    if (g <= gd1.g) {
        if (d >= gd1.d)
        { return 1; }
    }
    return 0;
}

int gd::operator<=(const gd& gd1) // comparaison d'un monome avec le monome courant
{
    if (g >= gd1.g) {
        if (d <= gd1.d)
        { return 1; }
    }
    return 0;
}

bool gd::operator<(const gd& gd1) const // comparaison d'un monome avec le monome courant
{
    if (g < gd1.g)
    { return true; }

    if (g == gd1.g) {
        if (d > gd1.d)
        { return true; }
    }

    return false;
}

gd& gd::init(long g1, long d1) // initialise avec 2  entiers
{
    affecte(g1, d1); // affectation
    return *this;
} // retourne l'objet courant

gd& gd::operator ()(long g1, long d1) // initialise avec un tableau de 2 entiers
{
    affecte(g1, d1); // affectation
    return *this;
} // retourne l'objet courant

gd inf(const gd& gd1, const gd& gd2)
// inf de 2 monomes, en entree 2 monomes par reference
// la fonction retourne
// un monome pour permettre un appel successif
{
    gd gdtemp(gd1);

    if (gd2.g > gd1.g)
    { gdtemp.g = gd2.g; }

    if (gd2.d < gd1.d)
    { gdtemp.d = gd2.d; }

    return gdtemp;
}

gd otimes(const gd& gd1, const gd& gd2)
// produit de 2 monomes, on traite les cas degeneres
// en entree 2 monomes par reference
// la fonction retourne
// un monome  pour permettre un appel successif
{
    gd temp;

    if (gd1.g == infty || gd2.g == infty) {
        temp.g = infty;
    } else if (gd1.g == _infty || gd2.g == _infty) {
        temp.g = _infty;
    } else {
        temp.g = gd1.g + gd2.g;
    }

    if (gd1.d == _infty || gd2.d == _infty) {
        temp.d = _infty;
    } else if (gd1.d == infty || gd2.d == infty) {
        temp.d = infty;
    } else {
        temp.d = gd1.d + gd2.d;
    }

    return temp;
}

gd frac(const gd& gd1, const gd& gd2)
// residuation de 2 monomes
// en entree 2 monomes par reference
// la fonction retourne
// un monome pour permettre un appel successif
{
    gd temp;
    switch (gd1.g) {
        case _infty:
            temp.g = _infty;
            break;
        case infty:
            if (gd2.g == infty)
            { temp.g = _infty; }
            else
            { temp.g = infty; }
            break;
        default
                :
            switch (gd2.g) {
                case infty :
                    temp.g = _infty;
                    break;
                case _infty :
                    temp.g = infty;
                    break;
                default
                        :
                    temp.g = gd1.g - gd2.g;
            }

    }
    switch (gd1.d) {
        case infty :
            temp.d = infty;
            break;
        case _infty  :
            if (gd2.d == _infty)
            { temp.d = infty; }
            else
            { temp.d = _infty; }
            break;
        default
                :
            switch (gd2.d) {
                case _infty :
                    temp.d = infty;
                    break;
                case infty :
                    temp.d = -infty;
                    break;
                default
                        :
                    temp.d = gd1.d - gd2.d;
            }
    }
    return (temp);
}

gd Dualfrac(const gd& gd1, const gd& gd2)
// residuation de 2 monomes
// en entree 2 monomes par reference
// la fonction retourne
// un monome pour permettre un appel successif
{
    gd temp;

    if (gd1.g == infty) {
        temp.g = infty;
    } else if (gd1.g == _infty) {
        temp.g = ((gd2.g == _infty) ? infty : _infty);
    } else {
        if (gd2.g == _infty) {
            temp.g = infty;
        } else if (gd2.g == infty) {
            temp.g = _infty;
        } else {
            temp.g = gd1.g - gd2.g;
        }
    }

    if (gd1.d == _infty) {
        temp.d = _infty;
    } else if (gd1.d == infty) {
        temp.d = ((gd2.d == infty) ? _infty : infty);
    } else {
        if (gd2.d == infty) {
            temp.d = _infty;
        } else if (gd2.d == _infty) {
            temp.d = infty;
        } else {
            temp.d = gd1.d - gd2.d;
        }
    }

    return temp;
}
