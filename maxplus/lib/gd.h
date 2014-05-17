#ifndef __GD_H__
#define __GD_H__

#include <stdlib.h>
#include <math.h>

/*
 * Classe pour exception
 */
class mem_limite
{
    public:
        int memoire;
        mem_limite(int i)
        {memoire = i ;}

};
class taille_incorrecte
{
    public:
        int erreur;
        taille_incorrecte(int i)
        {erreur = i ;}

};

#define infty 2147483647
#define _infty -2147483647

/*
 * Classe gd definissant un monome
 */
class gd
{
    private:
        long g;
        long d;
        void affecte(long, long);

    public:
        gd(void); // constructuer 0 : (g,d)=(+00,-00)
        gd(const gd&); // constructeur 1 : par recopie
        gd(long, long); // constructeur 2 : initialise avec 2 entiers

        // surdefinition de l'operateur egal
        gd& operator=(const gd&); // initialise avec un objet gd ,
        int operator!=(const gd&); // comparaison d'un monome avec le monome courant
        int operator==(const gd&); // comparaison d'un monome avec le monome courant
        int operator>=(const gd&); // comparaison d'un monome avec le monome courant
        int operator<=(const gd&); // comparaison d'un monome avec le monome courant
        bool operator<(const gd&) const; // comparaison d'un monome avec le monome courant

        gd& init(long, long); // initialise avec un tableau de 2 entiers
        gd& operator ()(long, long); // initialise avec un tableau de 2 entiers
        
        long getg(void) {return g;} // retourne g
        long getd(void) {return d;} // retourne d pour acces de l'exterieur

        friend gd inf(const gd& gd1, const gd& gd2);
        // inf de 2 monomes, en entree 2 monomes par reference
        // la fonction retourne
        // un monome pour permettre un appel successif

        friend gd otimes(const gd& gd1, const  gd& gd2);
        // produit de 2 monomes, on traite les cas degeneres
        // en entree 2 monomes par reference
        // la fonction retourne
        // un monome pour permettre un appel successif

        friend gd frac(const gd& gd1, const  gd& gd2);
        // residuation de 2 monomes
        // en entree 2 monomes par reference
        // la fonction retourne
        // un monome pour permettre un appel successif

        friend gd Dualfrac(const gd& gd1, const gd& gd2);

}; // fin de la definition de class gd


#endif

